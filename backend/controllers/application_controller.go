package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// expectedResponseWindow is the response time promised to candidates after each status change.
const expectedResponseWindow = 5 * 24 * time.Hour

// candidateFields are the only fields an applicant may set.
type candidateFields struct {
	FullName          string `json:"full_name" binding:"required,max=255"`
	Email             string `json:"email" binding:"required,email,max=255"`
	Phone             string `json:"phone" binding:"omitempty,max=50"`
	ResumeURL         string `json:"resume_url" binding:"required,url"`
	CoverLetter       string `json:"cover_letter" binding:"max=10000"`
	YearsOfExperience int    `json:"years_of_experience" binding:"min=0,max=60"`
	CurrentPosition   string `json:"current_position" binding:"max=255"`
	LinkedinURL       string `json:"linkedin_url" binding:"omitempty,url,max=255"`
	PortfolioURL      string `json:"portfolio_url" binding:"omitempty,url,max=255"`
	ReferralSource    string `json:"referral_source" binding:"max=255"`
	ReferredByName    string `json:"referred_by_name" binding:"max=255"`
	ReferredByEmail   string `json:"referred_by_email" binding:"omitempty,email,max=255"`
	ReferredByPhone   string `json:"referred_by_phone" binding:"max=50"`
}

func (f *candidateFields) toApplication(job *models.Job) models.Application {
	return models.Application{
		JobID:             &job.ID,
		CompanyID:         job.CompanyID,
		FullName:          f.FullName,
		Email:             strings.ToLower(f.Email),
		Phone:             f.Phone,
		ResumeURL:         f.ResumeURL,
		CoverLetter:       f.CoverLetter,
		YearsOfExperience: f.YearsOfExperience,
		CurrentPosition:   f.CurrentPosition,
		LinkedinURL:       f.LinkedinURL,
		PortfolioURL:      f.PortfolioURL,
		ReferralSource:    f.ReferralSource,
		ReferredByName:    f.ReferredByName,
		ReferredByEmail:   f.ReferredByEmail,
		ReferredByPhone:   f.ReferredByPhone,
		Status:            models.StatusPending,
		AppliedAt:         time.Now(),
	}
}

type submitApplicationRequest struct {
	JobID uuid.UUID `json:"job_id" binding:"required"`
	candidateFields
}

func SubmitApplication(c *gin.Context) {
	var req submitApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}

	var job models.Job
	err := config.DB.First(&job, "id = ?", req.JobID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}
	if err != nil {
		serverError(c, "Failed to load job")
		return
	}
	if job.Closed(time.Now().UTC()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Applications closed",
			"message": "This job is no longer accepting applications.",
		})
		return
	}

	app := req.toApplication(&job)
	if err := config.DB.Create(&app).Error; err != nil {
		serverError(c, "Failed to submit application")
		return
	}

	go services.ProcessApplicationCV(app.ID, app.ResumeURL, &job)
	go func() {
		if err := services.SendConfirmationEmail(app.Email, app.FullName, job.Title, app.ID.String()); err != nil {
			log.Printf("confirmation email to application %s: %v", app.ID, err)
		}
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Application submitted successfully!",
		"application": gin.H{"id": app.ID, "status": app.Status, "applied_at": app.AppliedAt},
	})
}

// GetApplications lists applications for the company's existing jobs; orphaned ones live in Find Candidates.
func GetApplications(c *gin.Context) {
	query := config.DB.Preload("Job").
		Omit("parsed_cv_text").
		Where("company_id = ? AND job_id IS NOT NULL", companyID(c)).
		Order("applied_at desc")

	if jobID := c.Query("job_id"); jobID != "" {
		query = query.Where("job_id = ?", jobID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if from, err := time.Parse("2006-01-02", c.Query("date_from")); err == nil {
		query = query.Where("applied_at >= ?", from)
	}
	if to, err := time.Parse("2006-01-02", c.Query("date_to")); err == nil {
		query = query.Where("applied_at < ?", to.AddDate(0, 0, 1))
	}

	var applications []models.Application
	if err := query.Find(&applications).Error; err != nil {
		serverError(c, "Failed to fetch applications")
		return
	}
	c.JSON(http.StatusOK, gin.H{"applications": applications})
}

func ShortlistApplication(c *gin.Context) {
	updateDecision(c, models.StatusShortlisted, services.SendShortlistEmail)
}

func RejectApplication(c *gin.Context) {
	updateDecision(c, models.StatusRejected, services.SendRejectionEmail)
}

func updateDecision(c *gin.Context, status string, notify func(to, name, jobTitle string) error) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	app, ok := findCompanyApplication(c, id)
	if !ok {
		return
	}

	now := time.Now()
	reviewer := adminID(c)
	oldStatus := app.Status
	app.Status = status
	app.ReviewedAt = &now
	app.ReviewedBy = &reviewer
	app.LastStatusUpdate = &now
	updates := map[string]any{
		"status": status, "reviewed_at": now, "reviewed_by": reviewer, "last_status_update": now,
	}
	if status == models.StatusShortlisted {
		app.ExpectedResponseDate = ptr(now.Add(expectedResponseWindow))
		updates["expected_response_date"] = app.ExpectedResponseDate
	}
	if err := config.DB.Model(app).Updates(updates).Error; err != nil {
		serverError(c, "Failed to update application")
		return
	}

	jobTitle := app.JobTitle()
	services.LogApplicationStatusChanged(app.CompanyID, reviewer, app.ID, app.FullName, jobTitle, oldStatus, status)
	go func() {
		if err := notify(app.Email, app.FullName, jobTitle); err != nil {
			log.Printf("%s email for application %s: %v", status, app.ID, err)
		}
		if err := services.SendStatusUpdateSMS(app.Phone, app.FullName, jobTitle, status); err != nil {
			log.Printf("%s SMS for application %s: %v", status, app.ID, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Application " + status, "application": app})
}

func DeleteApplication(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	app, ok := findCompanyApplication(c, id)
	if !ok {
		return
	}
	if err := config.DB.Delete(app).Error; err != nil {
		serverError(c, "Failed to delete application")
		return
	}

	company, admin, jobTitle := app.CompanyID, adminID(c), app.JobTitle()
	services.LogActivity(&company, &admin, "application_deleted", "application", &app.ID,
		"Application deleted: "+app.FullName+" for job: "+jobTitle,
		map[string]any{"applicant_name": app.FullName, "applicant_email": app.Email, "job_title": jobTitle, "status": app.Status})

	c.JSON(http.StatusOK, gin.H{"message": "Application deleted successfully"})
}

func BulkDeleteApplications(c *gin.Context) {
	var req struct {
		Status string `json:"status" binding:"required,oneof=pending cv_viewed shortlisted rejected"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}

	company, admin := companyID(c), adminID(c)
	var deleted []models.Application
	err := config.DB.Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}, {Name: "full_name"}, {Name: "email"}}}).
		Where("company_id = ? AND status = ?", company, req.Status).
		Delete(&deleted).Error
	if err != nil {
		serverError(c, "Failed to delete applications")
		return
	}

	if len(deleted) > 0 {
		names := make([]string, len(deleted))
		for i, app := range deleted {
			names[i] = app.FullName
		}
		services.LogActivity(&company, &admin, "applications_bulk_deleted", "application", nil,
			fmt.Sprintf("Bulk deleted %d %s application(s)", len(deleted), req.Status),
			map[string]any{"status": req.Status, "count": len(deleted), "applicants": names})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       fmt.Sprintf("Deleted %d application(s) with status '%s'", len(deleted), req.Status),
		"deleted_count": len(deleted),
		"total_found":   len(deleted),
	})
}

// TrackCVView records the first CV view and moves pending applications to cv_viewed.
func TrackCVView(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	app, ok := findCompanyApplication(c, id)
	if !ok {
		return
	}

	now := time.Now()
	viewer := adminID(c)
	updates := map[string]any{}
	if app.CVViewedAt == nil {
		app.CVViewedAt, app.CVViewedBy = &now, &viewer
		updates["cv_viewed_at"], updates["cv_viewed_by"] = now, viewer
	}
	oldStatus := app.Status
	if app.Status == models.StatusPending {
		app.Status = models.StatusCVViewed
		app.LastStatusUpdate = &now
		app.ExpectedResponseDate = ptr(now.Add(expectedResponseWindow))
		updates["status"], updates["last_status_update"], updates["expected_response_date"] = app.Status, now, app.ExpectedResponseDate
	}

	if len(updates) > 0 {
		if err := config.DB.Model(app).Updates(updates).Error; err != nil {
			serverError(c, "Failed to update application")
			return
		}
	}

	if oldStatus != app.Status {
		jobTitle := app.JobTitle()
		services.LogApplicationStatusChanged(app.CompanyID, viewer, app.ID, app.FullName, jobTitle, oldStatus, app.Status)
		go func() {
			if err := services.SendStatusUpdateSMS(app.Phone, app.FullName, jobTitle, app.Status); err != nil {
				log.Printf("cv_viewed SMS for application %s: %v", app.ID, err)
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{"message": "CV view tracked", "application": app})
}
