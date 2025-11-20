package controllers

import (
	"ats-backend/config"
	"ats-backend/models"
	"net/http"
	"time"

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

	// Get unread message count
	var unreadCount int64
	config.DB.Model(&models.Message{}).
		Where("application_id = ? AND sender_type = 'recruiter' AND is_read = false", application.ID).
		Count(&unreadCount)

	// Get status timeline/history
	statusHistory := []gin.H{
		{
			"status":    "pending",
			"label":     "Application Submitted",
			"timestamp": application.AppliedAt,
			"completed": application.Status != "pending",
		},
	}

	if application.CVViewedAt != nil {
		statusHistory = append(statusHistory, gin.H{
			"status":    "cv_viewed",
			"label":     "CV Viewed",
			"timestamp": application.CVViewedAt,
			"completed": true,
		})
	}

	if application.Status == "shortlisted" || application.Status == "rejected" {
		statusHistory = append(statusHistory, gin.H{
			"status":    application.Status,
			"label":     getStatusLabel(application.Status),
			"timestamp": application.ReviewedAt,
			"completed": true,
		})
	}

	// Calculate expected response date
	var expectedResponseDate *time.Time
	var expectedResponseDays int
	if application.ExpectedResponseDate != nil {
		expectedResponseDate = application.ExpectedResponseDate
		daysUntil := int(time.Until(*expectedResponseDate).Hours() / 24)
		if daysUntil > 0 {
			expectedResponseDays = daysUntil
		}
	} else if application.Status == "pending" || application.Status == "cv_viewed" {
		// Default: 5 days from last status update or applied date
		baseDate := application.AppliedAt
		if application.LastStatusUpdate != nil {
			baseDate = *application.LastStatusUpdate
		}
		defaultDate := baseDate.AddDate(0, 0, 5)
		expectedResponseDate = &defaultDate
		expectedResponseDays = int(time.Until(defaultDate).Hours() / 24)
	}

	// Return enhanced status information
	c.JSON(http.StatusOK, gin.H{
		"application": gin.H{
			"id":                    application.ID,
			"full_name":             application.FullName,
			"email":                 application.Email,
			"status":                application.Status,
			"status_label":          getStatusLabel(application.Status),
			"applied_at":            application.AppliedAt,
			"reviewed_at":           application.ReviewedAt,
			"cv_viewed_at":          application.CVViewedAt,
			"last_status_update":    application.LastStatusUpdate,
			"expected_response_date": expectedResponseDate,
			"expected_response_days": expectedResponseDays,
			"score":                 application.Score,
			"unread_messages":       unreadCount,
			"status_history":        statusHistory,
			"can_message":           true, // Candidates can always message
			"job": gin.H{
				"id":    application.Job.ID,
				"title": application.Job.Title,
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

// getStatusLabel returns a human-readable label for status
func getStatusLabel(status string) string {
	labels := map[string]string{
		"pending":            "Application Pending",
		"cv_viewed":          "CV Viewed",
		"under_review":       "Under Review",
		"shortlisted":        "Shortlisted",
		"rejected":           "Not Selected",
		"interview_scheduled": "Interview Scheduled",
		"decision_pending":   "Decision Pending",
	}
	if label, ok := labels[status]; ok {
		return label
	}
	return status
}




