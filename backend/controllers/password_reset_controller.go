package controllers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"ats-backend/config"
	"ats-backend/middleware"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const resetTokenTTL = time.Hour

var errInvalidResetToken = errors.New("invalid reset token")

func hashResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// ForgotPassword always answers the same way so it can't be used to discover registered emails.
func ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	ok := gin.H{"message": "If an account exists for that email, a reset link is on its way."}

	var admin models.Admin
	if err := config.DB.Where("lower(email) = ?", strings.ToLower(req.Email)).First(&admin).Error; err != nil {
		c.JSON(http.StatusOK, ok)
		return
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		serverError(c, "Failed to create reset link")
		return
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	record := models.PasswordResetToken{AdminID: admin.ID, TokenHash: hashResetToken(token), ExpiresAt: time.Now().Add(resetTokenTTL)}
	if err := config.DB.Create(&record).Error; err != nil {
		serverError(c, "Failed to create reset link")
		return
	}

	link := config.FrontendURL() + "/admin/reset-password?token=" + token
	go func() {
		if err := services.SendPasswordResetEmail(admin.Email, admin.Name, link); err != nil {
			log.Printf("password reset email for admin %s: %v", admin.ID, err)
		}
	}()
	c.JSON(http.StatusOK, ok)
}

// ResetPassword consumes a token, sets the new password and signs out every existing session.
func ResetPassword(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=8,max=72"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		serverError(c, "Failed to hash password")
		return
	}

	invalid := gin.H{"error": "This reset link is invalid or has expired. Please request a new one."}
	var adminID uuid.UUID
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		var record models.PasswordResetToken
		now := time.Now()
		// The row lock makes the token single-use even if two resets race.
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", hashResetToken(req.Token), now).
			First(&record).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errInvalidResetToken
		}
		if err != nil {
			return err
		}
		if err := tx.Model(&record).Update("used_at", now).Error; err != nil {
			return err
		}
		adminID = record.AdminID
		if err := tx.Model(&models.Admin{}).Where("id = ?", record.AdminID).
			// Receiving the reset email proves ownership, so it also verifies the address.
			Updates(map[string]any{
				"password_hash": string(hashed), "password_changed_at": now,
				"email_verified_at": gorm.Expr("coalesce(email_verified_at, ?)", now),
			}).Error; err != nil {
			return err
		}
		// Any other outstanding links for this admin are now stale.
		return tx.Where("admin_id = ? AND used_at IS NULL", record.AdminID).Delete(&models.PasswordResetToken{}).Error
	})
	if errors.Is(err, errInvalidResetToken) {
		c.JSON(http.StatusBadRequest, invalid)
		return
	}
	if err != nil {
		serverError(c, "Failed to reset password")
		return
	}
	middleware.ForgetAdminSessions(adminID)
	c.JSON(http.StatusOK, gin.H{"message": "Password updated. You can now log in."})
}
