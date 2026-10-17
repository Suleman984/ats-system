package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	verificationTTL      = 24 * time.Hour
	verificationCooldown = time.Minute
)

var errInvalidVerification = errors.New("invalid verification token")

// checkInbox is the only answer registration and resend give, so neither reveals whether an email
// already has an account.
var checkInbox = gin.H{"message": "Check your inbox for a link to verify your email and finish signing up."}

// sendVerificationEmail issues a fresh single-use link, unless one was sent in the last minute.
func sendVerificationEmail(admin *models.Admin) {
	var recent int64
	config.DB.Model(&models.EmailVerificationToken{}).
		Where("admin_id = ? AND created_at > ?", admin.ID, time.Now().Add(-verificationCooldown)).Count(&recent)
	if recent > 0 {
		return
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		log.Printf("verification token for admin %s: %v", admin.ID, err)
		return
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	record := models.EmailVerificationToken{AdminID: admin.ID, TokenHash: hashResetToken(token), ExpiresAt: time.Now().Add(verificationTTL)}
	if err := config.DB.Create(&record).Error; err != nil {
		log.Printf("verification token for admin %s: %v", admin.ID, err)
		return
	}
	link := config.FrontendURL() + "/admin/verify-email?token=" + token
	go func() {
		if err := services.SendVerificationEmail(admin.Email, admin.Name, link); err != nil {
			log.Printf("verification email for admin %s: %v", admin.ID, err)
		}
	}()
}

// handleExistingSignup runs when someone registers an email that already has an account.
func handleExistingSignup(email string) {
	var admin models.Admin
	if err := config.DB.Where("lower(email) = ?", email).First(&admin).Error; err != nil {
		log.Printf("sign-up for %s hit a duplicate but no admin has that email: %v", email, err)
		return
	}
	if admin.EmailVerifiedAt == nil {
		sendVerificationEmail(&admin)
		return
	}
	go func() {
		if err := services.SendAccountExistsEmail(admin.Email, admin.Name); err != nil {
			log.Printf("account-exists email for admin %s: %v", admin.ID, err)
		}
	}()
}

func ResendVerification(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	var admin models.Admin
	err := config.DB.Where("lower(email) = ? AND email_verified_at IS NULL", strings.ToLower(req.Email)).First(&admin).Error
	if err == nil {
		sendVerificationEmail(&admin)
	}
	c.JSON(http.StatusOK, checkInbox)
}

// VerifyEmail consumes a link, marks the email verified and signs the admin in.
func VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	var admin models.Admin
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var record models.EmailVerificationToken
		now := time.Now()
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", hashResetToken(req.Token), now).
			First(&record).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errInvalidVerification
		}
		if err != nil {
			return err
		}
		if err := tx.Model(&record).Update("used_at", now).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Admin{}).Where("id = ? AND email_verified_at IS NULL", record.AdminID).
			Update("email_verified_at", now).Error; err != nil {
			return err
		}
		if err := tx.Where("admin_id = ? AND used_at IS NULL", record.AdminID).Delete(&models.EmailVerificationToken{}).Error; err != nil {
			return err
		}
		return tx.First(&admin, "id = ?", record.AdminID).Error
	})
	if errors.Is(err, errInvalidVerification) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This verification link is invalid or has expired. Request a new one from the login page."})
		return
	}
	if err != nil {
		serverError(c, "Failed to verify email")
		return
	}
	respondWithAdminToken(c, http.StatusOK, &admin, "Email verified")
}
