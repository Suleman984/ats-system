package controllers

import (
	"ats-backend/config"
	"ats-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CheckApplicationStatusRequest for candidate portal
type CheckApplicationStatusRequest struct {
	Email         string `json:"email" binding:"required,email"`
	ApplicationID string `json:"application_id" binding:"required"`
}

// GetApplicationStatus returns application status for candidates (public endpoint)
func GetApplicationStatus(c *gin.Context) {
	var req CheckApplicationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var application models.Application
	err := config.DB.Where("id = ? AND email = ?", req.ApplicationID, req.Email).
		Preload("Job").
		First(&application).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found. Please check your email and application ID.",
		})
		return
	}

	// Return status information
	c.JSON(http.StatusOK, gin.H{
		"application": gin.H{
			"id":              application.ID,
			"full_name":       application.FullName,
			"email":           application.Email,
			"status":          application.Status,
			"applied_at":      application.AppliedAt,
			"reviewed_at":     application.ReviewedAt,
			"score":           application.Score,
			"job": gin.H{
				"id":          application.Job.ID,
				"title":       application.Job.Title,
				"company_name": "", // Will be populated if needed
			},
		},
	})
}

// GetApplicationStatusByEmail returns all applications for an email (public endpoint)
func GetApplicationStatusByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	var applications []models.Application
	err := config.DB.Where("email = ?", email).
		Preload("Job").
		Order("applied_at DESC").
		Find(&applications).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
		return
	}

	// Return simplified application info (no sensitive data)
	results := []gin.H{}
	for _, app := range applications {
		results = append(results, gin.H{
			"id":         app.ID,
			"full_name":  app.FullName,
			"status":     app.Status,
			"applied_at": app.AppliedAt,
			"reviewed_at": app.ReviewedAt,
			"job": gin.H{
				"id":    app.Job.ID,
				"title": app.Job.Title,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"applications": results,
		"count":       len(results),
	})
}




