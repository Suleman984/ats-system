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
	"ats-backend/middleware"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const inviteTTL = 7 * 24 * time.Hour

var errInvalidInvite = errors.New("invalid invite")

// GetTeam lists the company's members and open invites.
func GetTeam(c *gin.Context) {
	company := companyID(c)
	var members []models.Admin
	if err := config.DB.Where("company_id = ?", company).Order("created_at").Find(&members).Error; err != nil {
		serverError(c, "Failed to load team")
		return
	}
	var invites []models.TeamInvite
	if err := config.DB.Where("company_id = ? AND accepted_at IS NULL AND expires_at > now()", company).
		Order("created_at desc").Find(&invites).Error; err != nil {
		serverError(c, "Failed to load invites")
		return
	}
	c.JSON(http.StatusOK, gin.H{"members": members, "invites": invites})
}

func InviteTeamMember(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email,max=255"`
		Role  string `json:"role" binding:"required,oneof=owner recruiter"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	company, inviter := companyID(c), adminID(c)
	email := strings.ToLower(req.Email)

	var existing int64
	config.DB.Model(&models.Admin{}).Where("lower(email) = ?", email).Count(&existing)
	if existing > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "That email already has an account"})
		return
	}
	// Re-inviting replaces the old link rather than counting twice against the seat limit.
	if err := config.DB.Where("company_id = ? AND lower(email) = ? AND accepted_at IS NULL", company, email).
		Delete(&models.TeamInvite{}).Error; err != nil {
		serverError(c, "Failed to create invite")
		return
	}
	if !planCheck(c, services.CheckCanAddMember(company)) {
		return
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		serverError(c, "Failed to create invite")
		return
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	invite := models.TeamInvite{
		CompanyID: company, Email: email, Role: req.Role, TokenHash: hashResetToken(token),
		InvitedBy: &inviter, ExpiresAt: time.Now().Add(inviteTTL),
	}
	if err := config.DB.Create(&invite).Error; err != nil {
		serverError(c, "Failed to create invite")
		return
	}

	var meta struct{ CompanyName, InviterName string }
	config.DB.Raw(`select c.company_name, a.name as inviter_name from companies c join admins a on a.company_id = c.id
		where c.id = ? and a.id = ?`, company, inviter).Scan(&meta)
	link := config.FrontendURL() + "/admin/accept-invite?token=" + token
	go func() {
		if err := services.SendTeamInviteEmail(email, meta.InviterName, meta.CompanyName, req.Role, link); err != nil {
			log.Printf("team invite email for %s: %v", invite.ID, err)
		}
	}()

	services.LogActivity(&company, &inviter, "team_member_invited", "admin", nil,
		"Invited "+email+" as "+req.Role, map[string]any{"email": email, "role": req.Role})
	c.JSON(http.StatusCreated, gin.H{"message": "Invite sent", "invite": invite})
}

func RevokeTeamInvite(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	res := config.DB.Where("id = ? AND company_id = ? AND accepted_at IS NULL", id, companyID(c)).Delete(&models.TeamInvite{})
	if res.Error != nil {
		serverError(c, "Failed to revoke invite")
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invite not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invite revoked"})
}

// UpdateTeamMemberRole changes a teammate's role; the company always keeps at least one owner.
func UpdateTeamMemberRole(c *gin.Context) {
	var req struct {
		Role string `json:"role" binding:"required,oneof=owner recruiter"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	member, ok := loadTeamMember(c)
	if !ok {
		return
	}
	if member.Role == models.RoleOwner && req.Role != models.RoleOwner && !hasAnotherOwner(member) {
		c.JSON(http.StatusConflict, gin.H{"error": "The company needs at least one owner"})
		return
	}
	if err := config.DB.Model(member).Update("role", req.Role).Error; err != nil {
		serverError(c, "Failed to update role")
		return
	}
	middleware.ForgetAdminSessions(member.ID)
	company, admin := companyID(c), adminID(c)
	services.LogActivity(&company, &admin, "team_member_role_changed", "admin", &member.ID,
		member.Name+" is now "+req.Role, map[string]any{"email": member.Email, "role": req.Role})
	c.JSON(http.StatusOK, gin.H{"message": "Role updated", "member": member})
}

func RemoveTeamMember(c *gin.Context) {
	member, ok := loadTeamMember(c)
	if !ok {
		return
	}
	if member.ID == adminID(c) {
		c.JSON(http.StatusConflict, gin.H{"error": "You can't remove yourself"})
		return
	}
	if member.Role == models.RoleOwner && !hasAnotherOwner(member) {
		c.JSON(http.StatusConflict, gin.H{"error": "The company needs at least one owner"})
		return
	}
	if err := config.DB.Delete(member).Error; err != nil {
		serverError(c, "Failed to remove member")
		return
	}
	middleware.ForgetAdminSessions(member.ID)
	company, admin := companyID(c), adminID(c)
	services.LogActivity(&company, &admin, "team_member_removed", "admin", &member.ID,
		"Removed "+member.Name+" from the team", map[string]any{"email": member.Email})
	c.JSON(http.StatusOK, gin.H{"message": "Member removed"})
}

func loadTeamMember(c *gin.Context) (*models.Admin, bool) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return nil, false
	}
	var member models.Admin
	err := config.DB.Where("id = ? AND company_id = ?", id, companyID(c)).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return nil, false
	}
	if err != nil {
		serverError(c, "Failed to load team member")
		return nil, false
	}
	return &member, true
}

func hasAnotherOwner(member *models.Admin) bool {
	var n int64
	config.DB.Model(&models.Admin{}).
		Where("company_id = ? AND role = ? AND id <> ?", member.CompanyID, models.RoleOwner, member.ID).Count(&n)
	return n > 0
}

// GetInvite lets the accept page show who is inviting whom, without revealing anything else.
func GetInvite(c *gin.Context) {
	var invite models.TeamInvite
	err := config.DB.Where("token_hash = ? AND accepted_at IS NULL AND expires_at > now()", hashResetToken(c.Query("token"))).
		First(&invite).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "This invite is invalid or has expired"})
		return
	}
	var company models.Company
	config.DB.Select("company_name").First(&company, "id = ?", invite.CompanyID)
	c.JSON(http.StatusOK, gin.H{"email": invite.Email, "role": invite.Role, "company_name": company.CompanyName})
}

// AcceptInvite creates the invited admin's account and signs them in.
func AcceptInvite(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Name     string `json:"name" binding:"required,max=255"`
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

	var admin models.Admin
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		var invite models.TeamInvite
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ? AND accepted_at IS NULL AND expires_at > now()", hashResetToken(req.Token)).
			First(&invite).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errInvalidInvite
		}
		if err != nil {
			return err
		}
		// The invite email already proved the address.
		now := time.Now()
		admin = models.Admin{
			CompanyID: invite.CompanyID, Name: req.Name, Email: invite.Email, PasswordHash: string(hashed),
			Role: invite.Role, EmailVerifiedAt: &now,
		}
		if err := tx.Create(&admin).Error; err != nil {
			return err
		}
		return tx.Model(&invite).Update("accepted_at", time.Now()).Error
	})
	if errors.Is(err, errInvalidInvite) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This invite is invalid or has expired. Ask your team owner for a new one."})
		return
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		c.JSON(http.StatusConflict, gin.H{"error": "That email already has an account"})
		return
	}
	if err != nil {
		serverError(c, "Failed to accept invite")
		return
	}

	company := admin.CompanyID
	services.LogActivity(&company, &admin.ID, "team_member_joined", "admin", &admin.ID,
		admin.Name+" joined the team as "+admin.Role, map[string]any{"email": admin.Email})
	respondWithAdminToken(c, http.StatusCreated, &admin, "Welcome to the team")
}

// GetMe returns the signed-in admin, so the UI can refresh role and name.
func GetMe(c *gin.Context) {
	var admin models.Admin
	if err := config.DB.First(&admin, "id = ?", adminID(c)).Error; err != nil {
		serverError(c, "Failed to load account")
		return
	}
	c.JSON(http.StatusOK, gin.H{"admin": adminJSON(&admin)})
}
