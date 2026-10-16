package controllers

import (
	"errors"
	"net/http"
	"strings"

	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type manualCandidateRequest struct {
	JobID  uuid.UUID `json:"job_id" binding:"required"`
	Status string    `json:"status" binding:"omitempty,oneof=pending cv_viewed shortlisted rejected"`
	Notes  string    `json:"notes" binding:"max=5000"`
	candidateFields
}

// AddManualCandidate lets recruiters add a candidate who didn't apply through the portal.
func AddManualCandidate(c *gin.Context) {
	var req manualCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	job, ok := loadCompanyJob(c, req.JobID)
	if !ok {
		return
	}

	var existing models.Application
	err := config.DB.Select("id").Where("job_id = ? AND email = ?", job.ID, strings.ToLower(req.Email)).First(&existing).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Candidate already applied for this job", "application_id": existing.ID})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		serverError(c, "Failed to check existing applications")
		return
	}

	app := req.toApplication(job)
	if req.Status != "" {
		app.Status = req.Status
	}
	if err := config.DB.Create(&app).Error; err != nil {
		serverError(c, "Failed to add candidate")
		return
	}

	go services.ProcessApplicationCV(app.ID, app.ResumeURL, job)

	company, admin := job.CompanyID, adminID(c)
	services.LogActivity(&company, &admin, "manual_candidate_added", "application", &app.ID,
		"Manually added candidate: "+app.FullName+" for job: "+job.Title,
		map[string]any{"candidate_name": app.FullName, "candidate_email": app.Email, "job_title": job.Title, "notes": req.Notes})

	c.JSON(http.StatusCreated, gin.H{"message": "Candidate added successfully", "application": app})
}
