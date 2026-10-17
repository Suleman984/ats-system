package controllers

import (
	"errors"
	"net/http"
	"strings"

	"ats-backend/config"
	"ats-backend/middleware"
	"ats-backend/models"
	"ats-backend/services"
	"ats-backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SuperAdminLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}

	var superAdmin models.SuperAdmin
	err := config.DB.Where("lower(email) = ?", strings.ToLower(req.Email)).First(&superAdmin).Error
	hash := dummyHash
	if err == nil {
		hash = []byte(superAdmin.PasswordHash)
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(req.Password)) != nil || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := utils.GenerateSuperAdminJWT(superAdmin.ID)
	if err != nil {
		serverError(c, "Failed to generate token")
		return
	}
	middleware.SetSessionCookie(c, middleware.SuperAdminCookie, token)
	c.JSON(http.StatusOK, gin.H{
		"super_admin": gin.H{"id": superAdmin.ID, "name": superAdmin.Name, "email": superAdmin.Email},
	})
}

type platformStats struct {
	TotalCompanies          int64 `json:"total_companies"`
	ActiveCompanies         int64 `json:"active_companies"`
	TotalJobs               int64 `json:"total_jobs"`
	OpenJobs                int64 `json:"open_jobs"`
	TotalApplications       int64 `json:"total_applications"`
	PendingApplications     int64 `json:"pending_applications"`
	ShortlistedApplications int64 `json:"shortlisted_applications"`
	TotalAdmins             int64 `json:"total_admins"`
}

func GetSuperAdminStats(c *gin.Context) {
	var stats platformStats
	err := config.DB.Raw(`
		select
			(select count(*) from companies) as total_companies,
			(select count(*) from companies where subscription_status <> 'suspended') as active_companies,
			(select count(*) from jobs) as total_jobs,
			(select count(*) from jobs where status = 'open') as open_jobs,
			(select count(*) from applications) as total_applications,
			(select count(*) from applications where status = 'pending') as pending_applications,
			(select count(*) from applications where status = 'shortlisted') as shortlisted_applications,
			(select count(*) from admins) as total_admins
	`).Scan(&stats).Error
	if err != nil {
		serverError(c, "Failed to load stats")
		return
	}
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

type companyWithStats struct {
	models.Company
	JobCount         int64 `json:"job_count"`
	ApplicationCount int64 `json:"application_count"`
}

func GetAllCompanies(c *gin.Context) {
	// Raw scans leave a nil slice when nothing matches, which would serialize as null.
	companies := []companyWithStats{}
	err := config.DB.Raw(`
		select c.*,
			(select count(*) from jobs j where j.company_id = c.id) as job_count,
			(select count(*) from applications a where a.company_id = c.id) as application_count
		from companies c
		order by c.created_at desc
	`).Scan(&companies).Error
	if err != nil {
		serverError(c, "Failed to fetch companies")
		return
	}
	c.JSON(http.StatusOK, gin.H{"companies": companies})
}

// UpdateCompanySubscription lets a super admin suspend/reactivate a company or change its plan tier.
func UpdateCompanySubscription(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		SubscriptionStatus *string `json:"subscription_status" binding:"omitempty,oneof=trial active suspended"`
		SubscriptionTier   *string `json:"subscription_tier" binding:"omitempty,oneof=starter professional enterprise"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}

	var company models.Company
	err := config.DB.First(&company, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}
	if err != nil {
		serverError(c, "Failed to load company")
		return
	}

	before := map[string]string{"subscription_status": company.SubscriptionStatus, "subscription_tier": company.SubscriptionTier}
	updates := map[string]any{}
	if req.SubscriptionStatus != nil {
		updates["subscription_status"] = *req.SubscriptionStatus
	}
	if req.SubscriptionTier != nil {
		updates["subscription_tier"] = *req.SubscriptionTier
	}
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nothing to update"})
		return
	}
	if err := config.DB.Model(&company).Updates(updates).Error; err != nil {
		serverError(c, "Failed to update company")
		return
	}
	middleware.ForgetCompanySessions(company.ID)

	services.LogActivity(&company.ID, nil, "company_subscription_updated", "company", &company.ID,
		"Subscription updated for "+company.CompanyName, map[string]any{"before": before, "after": updates})
	c.JSON(http.StatusOK, gin.H{"message": "Company updated", "company": company})
}
