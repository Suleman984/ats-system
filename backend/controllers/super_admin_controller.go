package controllers

import (
	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type SuperAdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// SuperAdminLogin authenticates super admin and returns JWT token
func SuperAdminLogin(c *gin.Context) {
	var req SuperAdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var superAdmin models.SuperAdmin
	if err := config.DB.Where("email = ?", req.Email).First(&superAdmin).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(superAdmin.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token (use "super_admin" as company_id for super admin)
	token, err := utils.GenerateSuperAdminJWT(superAdmin.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"super_admin": gin.H{
			"id":    superAdmin.ID,
			"name":  superAdmin.Name,
			"email": superAdmin.Email,
		},
	})
}

// GetSuperAdminStats returns platform-wide statistics
func GetSuperAdminStats(c *gin.Context) {
	var stats struct {
		TotalCompanies      int64 `json:"total_companies"`
		ActiveCompanies     int64 `json:"active_companies"`
		TotalJobs           int64 `json:"total_jobs"`
		OpenJobs            int64 `json:"open_jobs"`
		TotalApplications   int64 `json:"total_applications"`
		PendingApplications int64 `json:"pending_applications"`
		ShortlistedApplications int64 `json:"shortlisted_applications"`
		TotalAdmins         int64 `json:"total_admins"`
	}

	// Get company stats
	config.DB.Model(&models.Company{}).Count(&stats.TotalCompanies)
	config.DB.Model(&models.Company{}).Where("subscription_status = ?", "active").Count(&stats.ActiveCompanies)

	// Get job stats
	config.DB.Model(&models.Job{}).Count(&stats.TotalJobs)
	config.DB.Model(&models.Job{}).Where("status = ?", "open").Count(&stats.OpenJobs)

	// Get application stats
	config.DB.Model(&models.Application{}).Count(&stats.TotalApplications)
	config.DB.Model(&models.Application{}).Where("status = ?", "pending").Count(&stats.PendingApplications)
	config.DB.Model(&models.Application{}).Where("status = ?", "shortlisted").Count(&stats.ShortlistedApplications)

	// Get admin stats
	config.DB.Model(&models.Admin{}).Count(&stats.TotalAdmins)

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// GetAllCompanies returns all companies with their stats
func GetAllCompanies(c *gin.Context) {
	var companies []models.Company
	if err := config.DB.Find(&companies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch companies"})
		return
	}

	// Get stats for each company
	type CompanyWithStats struct {
		models.Company
		JobCount        int64 `json:"job_count"`
		ApplicationCount int64 `json:"application_count"`
	}

	var companiesWithStats []CompanyWithStats
	for _, company := range companies {
		var jobCount, appCount int64
		config.DB.Model(&models.Job{}).Where("company_id = ?", company.ID).Count(&jobCount)
		config.DB.Model(&models.Application{}).
			Joins("JOIN jobs ON jobs.id = applications.job_id").
			Where("jobs.company_id = ?", company.ID).
			Count(&appCount)

		companiesWithStats = append(companiesWithStats, CompanyWithStats{
			Company:          company,
			JobCount:         jobCount,
			ApplicationCount: appCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{"companies": companiesWithStats})
}

