package controllers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"ats-backend/config"
	"ats-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var statusLabels = map[string]string{
	models.StatusPending:     "Application Pending",
	models.StatusCVViewed:    "CV Viewed",
	models.StatusShortlisted: "Shortlisted",
	models.StatusRejected:    "Not Selected",
}

func statusLabel(status string) string {
	if label, ok := statusLabels[status]; ok {
		return label
	}
	return status
}

// findCandidateApplication authenticates a candidate by application ID plus email.
func findCandidateApplication(c *gin.Context, applicationID uuid.UUID, email string) (*models.Application, bool) {
	var app models.Application
	err := config.DB.Preload("Job").Omit("parsed_cv_text").
		Where("id = ? AND lower(email) = ?", applicationID, strings.ToLower(email)).
		First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found. Please check your email and application ID."})
		return nil, false
	}
	if err != nil {
		serverError(c, "Failed to load application")
		return nil, false
	}
	return &app, true
}

func GetApplicationStatus(c *gin.Context) {
	var req struct {
		Email         string    `json:"email" binding:"required,email"`
		ApplicationID uuid.UUID `json:"application_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	app, ok := findCandidateApplication(c, req.ApplicationID, req.Email)
	if !ok {
		return
	}

	var unread int64
	config.DB.Model(&models.Message{}).
		Where("application_id = ? AND sender_type = ? AND is_read = false", app.ID, models.SenderRecruiter).
		Count(&unread)

	history := []gin.H{{
		"status": models.StatusPending, "label": "Application Submitted",
		"timestamp": app.AppliedAt, "completed": app.Status != models.StatusPending,
	}}
	if app.CVViewedAt != nil {
		history = append(history, gin.H{"status": models.StatusCVViewed, "label": "CV Viewed", "timestamp": app.CVViewedAt, "completed": true})
	}
	if app.Status == models.StatusShortlisted || app.Status == models.StatusRejected {
		history = append(history, gin.H{"status": app.Status, "label": statusLabel(app.Status), "timestamp": app.ReviewedAt, "completed": true})
	}

	expected := app.ExpectedResponseDate
	if expected == nil && (app.Status == models.StatusPending || app.Status == models.StatusCVViewed) {
		base := app.AppliedAt
		if app.LastStatusUpdate != nil {
			base = *app.LastStatusUpdate
		}
		expected = ptr(base.Add(expectedResponseWindow))
	}
	expectedDays := 0
	if expected != nil {
		expectedDays = max(0, int(time.Until(*expected).Hours()/24))
	}

	job := gin.H{"id": nil, "title": app.JobTitle()}
	if app.Job != nil {
		job["id"] = app.Job.ID
	}

	c.JSON(http.StatusOK, gin.H{"application": gin.H{
		"id":                     app.ID,
		"full_name":              app.FullName,
		"email":                  app.Email,
		"status":                 app.Status,
		"status_label":           statusLabel(app.Status),
		"applied_at":             app.AppliedAt,
		"reviewed_at":            app.ReviewedAt,
		"cv_viewed_at":           app.CVViewedAt,
		"last_status_update":     app.LastStatusUpdate,
		"expected_response_date": expected,
		"expected_response_days": expectedDays,
		"score":                  app.Score,
		"unread_messages":        unread,
		"status_history":         history,
		"can_message":            true,
		"job":                    job,
	}})
}
