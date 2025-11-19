package controllers

import (
	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SubmitApplication handles job application submission
func SubmitApplication(c *gin.Context) {
	var application models.Application
	if err := c.ShouldBindJSON(&application); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if job exists and is open
	var job models.Job
	jobID := application.JobID
	if jobID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}
	if err := config.DB.First(&job, "id = ?", *jobID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	// Check if job is closed
	if job.Status != "open" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Applications closed",
			"message": "This job posting is currently closed and not accepting applications.",
		})
		return
	}

	// Check if deadline has passed
	if time.Now().After(job.Deadline.Time) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Applications closed",
			"message": "Unfortunately, the deadline for this position has passed. Better luck next time!",
		})
		return
	}

	// Set application timestamp
	application.AppliedAt = time.Now()
	application.Status = "pending"
	// Set company_id from job (for tracking even if job is deleted later)
	application.CompanyID = job.CompanyID

	// Save to database
	if err := config.DB.Create(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit application"})
		return
	}

	// Parse and store CV text for future searching (async)
	go func() {
		if application.ResumeURL != "" {
			log.Printf("Parsing CV text for application %s", application.ID.String())
			cvText, err := services.ExtractTextFromURL(application.ResumeURL)
			if err == nil && len(cvText) > 50 {
				application.ParsedCVText = &cvText
				config.DB.Model(&application).Update("parsed_cv_text", cvText)
				log.Printf("CV text parsed and stored for %s (%d characters)", application.FullName, len(cvText))
			} else {
				log.Printf("Failed to parse CV text for %s: %v", application.Email, err)
			}
		}
	}()

	// Automatically analyze CV and calculate score (async)
	// Note: We only calculate score, admin decides whether to shortlist
	go func() {
		log.Printf("Starting CV analysis for application %s (Job: %s)", application.ID.String(), job.Title)
		
		var criteria services.Criteria
		
		// Try to use job's shortlist criteria if available
		if job.ShortlistCriteria != nil && *job.ShortlistCriteria != "" {
			log.Printf("Using job's shortlist criteria for application %s", application.ID.String())
			if err := json.Unmarshal([]byte(*job.ShortlistCriteria), &criteria); err != nil {
				log.Printf("WARNING: Failed to parse job's shortlist criteria for application %s: %v. Using default criteria.", application.ID.String(), err)
				// Fall through to use default criteria
			} else {
				criteria.JobDescription = job.Description
				criteria.JobRequirements = job.Requirements
			}
		} else {
			log.Printf("No shortlist criteria set for job '%s'. Using default criteria for basic analysis.", job.Title)
			// Use default criteria: analyze based on job description and requirements
			criteria = services.Criteria{
				RequiredSkills:      []string{}, // Will extract from job description
				MinExperience:       0,          // Will extract from job description
				RequiredLanguages:   []string{},
				MatchJobDescription: true,
				JobDescription:      job.Description,
				JobRequirements:     job.Requirements,
			}
		}
		
		// Analyze CV
		log.Printf("Analyzing CV for %s (Email: %s)", application.FullName, application.Email)
		analysisResult, err := services.MatchCVFromURL(application.ResumeURL, criteria, job.Title)
		if err != nil {
			log.Printf("ERROR: Failed to auto-analyze CV for %s (Email: %s, Application ID: %s): %v", 
				application.FullName, application.Email, application.ID.String(), err)
			log.Printf("CV URL: %s", application.ResumeURL)
			return
		}
		
		// Update application with score (but don't auto-shortlist)
		application.Score = analysisResult.MatchScore
		analysisJSON, _ := json.Marshal(analysisResult)
		analysisJSONStr := string(analysisJSON)
		application.AnalysisResult = &analysisJSONStr
		
		// Save updated application with score
		if err := config.DB.Save(&application).Error; err != nil {
			log.Printf("ERROR: Failed to save analysis result for application %s: %v", application.ID.String(), err)
			return
		}
		
		log.Printf("SUCCESS: CV analyzed for %s (Email: %s, Application ID: %s): Score=%d%% (Admin can review and shortlist manually)", 
			application.FullName, application.Email, application.ID.String(), analysisResult.MatchScore)
	}()

	// Send confirmation email (async with error logging)
	go func() {
		log.Printf("Sending confirmation email to %s for job %s", application.Email, job.Title)
		if err := services.SendConfirmationEmail(application.Email, application.FullName, job.Title); err != nil {
			log.Printf("ERROR: Failed to send confirmation email to %s: %v", application.Email, err)
		} else {
			log.Printf("SUCCESS: Confirmation email sent to %s", application.Email)
		}
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Application submitted successfully!",
		"application": application,
	})
}

// GetApplications returns all applications for company's jobs
func GetApplications(c *gin.Context) {
	companyIDVal, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in token"})
		return
	}
	companyID, ok := companyIDVal.(string)
	if !ok || companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	var applications []models.Application

	// Get all applications for this company, including those whose jobs were deleted
	// Strategy:
	// 1. Get applications where job exists and belongs to company
	// 2. If company_id column exists in applications, also get applications with NULL job_id for this company
	// Note: After running ADD_COMPANY_ID_TO_APPLICATIONS.sql, this will work perfectly
	query := config.DB.Table("applications").
		Select("applications.*").
		Joins("LEFT JOIN jobs ON jobs.id = applications.job_id").
		Where("jobs.company_id = ?", companyID).
		Preload("Job")
	
	// If company_id column exists, also include applications with deleted jobs
	// This will be handled automatically after migration adds company_id column

	// Filter by job_id if provided
	if jobID := c.Query("job_id"); jobID != "" {
		query = query.Where("applications.job_id = ?", jobID)
	}

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("applications.status = ?", status)
	}

	// Filter by date range (applied_at date)
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		// Parse date and filter by date (ignoring time)
		if dateFromTime, err := time.Parse("2006-01-02", dateFrom); err == nil {
			// Use PostgreSQL DATE() function to extract date part and compare
			// This ensures timezone-independent date comparison
			query = query.Where("DATE(applications.applied_at) >= ?", dateFromTime.Format("2006-01-02"))
			log.Printf("Filtering applications from date: %s", dateFromTime.Format("2006-01-02"))
		} else {
			log.Printf("Invalid date_from format: %s, error: %v", dateFrom, err)
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		// Parse date and filter by date (ignoring time)
		if dateToTime, err := time.Parse("2006-01-02", dateTo); err == nil {
			// Use PostgreSQL DATE() function to extract date part and compare
			query = query.Where("DATE(applications.applied_at) <= ?", dateToTime.Format("2006-01-02"))
			log.Printf("Filtering applications to date: %s", dateToTime.Format("2006-01-02"))
		} else {
			log.Printf("Invalid date_to format: %s, error: %v", dateTo, err)
		}
	}

	if err := query.Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"applications": applications})
}

// ShortlistApplication marks an application as shortlisted
func ShortlistApplication(c *gin.Context) {
	applicationID := c.Param("id")
	companyIDVal, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in token"})
		return
	}
	companyID, ok := companyIDVal.(string)
	if !ok || companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var application models.Application
	// Verify application belongs to company (even if job is deleted)
	err := config.DB.Table("applications").
		Select("applications.*").
		Joins("LEFT JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.id = ? AND (applications.company_id = ? OR jobs.company_id = ?)", applicationID, companyID, companyID).
		Preload("Job").
		First(&application).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	// Store old status for logging
	oldStatus := application.Status

	// Update status
	now := time.Now()
	application.Status = "shortlisted"
	application.ReviewedAt = &now

	if err := config.DB.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application"})
		return
	}

	// Log application shortlisting
	companyUUID, _ := uuid.Parse(companyID)
	adminIDVal, _ := c.Get("admin_id")
	adminIDStr, _ := adminIDVal.(string)
	adminUUID, _ := uuid.Parse(adminIDStr)
	applicationUUID, _ := uuid.Parse(applicationID)
	jobTitle := "Unknown Job (Job Deleted)"
	if application.JobID != nil && application.Job.ID != uuid.Nil {
		jobTitle = application.Job.Title
	}

	services.LogApplicationStatusChanged(companyUUID, adminUUID, applicationUUID, application.FullName, jobTitle, oldStatus, "shortlisted")

	// Send shortlist email (async with error logging)
	go func() {
		log.Printf("Sending shortlist email to %s for application %s", application.Email, applicationID)
		if err := services.SendShortlistEmail(application.Email, application.FullName, jobTitle); err != nil {
			log.Printf("ERROR: Failed to send shortlist email to %s: %v", application.Email, err)
		} else {
			log.Printf("SUCCESS: Shortlist email sent to %s", application.Email)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message":     "Application shortlisted successfully",
		"application": application,
	})
}

// RejectApplication marks an application as rejected
func RejectApplication(c *gin.Context) {
	applicationID := c.Param("id")
	companyIDVal, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in token"})
		return
	}
	companyID, ok := companyIDVal.(string)
	if !ok || companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var application models.Application
	// Verify application belongs to company (even if job is deleted)
	err := config.DB.Table("applications").
		Select("applications.*").
		Joins("LEFT JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.id = ? AND (applications.company_id = ? OR jobs.company_id = ?)", applicationID, companyID, companyID).
		Preload("Job").
		First(&application).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	// Store old status for logging
	oldStatus := application.Status

	now := time.Now()
	application.Status = "rejected"
	application.ReviewedAt = &now

	if err := config.DB.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application"})
		return
	}

	// Log application rejection
	companyUUID, _ := uuid.Parse(companyID)
	adminIDVal, _ := c.Get("admin_id")
	adminIDStr, _ := adminIDVal.(string)
	adminUUID, _ := uuid.Parse(adminIDStr)
	applicationUUID, _ := uuid.Parse(applicationID)
	jobTitle := "Unknown Job (Job Deleted)"
	if application.JobID != nil && application.Job.ID != uuid.Nil {
		jobTitle = application.Job.Title
	}

	services.LogApplicationStatusChanged(companyUUID, adminUUID, applicationUUID, application.FullName, jobTitle, oldStatus, "rejected")

	// Send rejection email (async with error logging)
	go func() {
		log.Printf("Sending rejection email to %s", application.Email)
		if err := services.SendRejectionEmail(application.Email, application.FullName, jobTitle); err != nil {
			log.Printf("ERROR: Failed to send rejection email to %s: %v", application.Email, err)
		} else {
			log.Printf("SUCCESS: Rejection email sent to %s", application.Email)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Application rejected",
	})
}

// DeleteApplication deletes a single application
func DeleteApplication(c *gin.Context) {
	applicationID := c.Param("id")
	companyIDVal, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in token"})
		return
	}
	companyID, ok := companyIDVal.(string)
	if !ok || companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var application models.Application
	// Verify application belongs to company (even if job is deleted)
	err := config.DB.Table("applications").
		Select("applications.*").
		Joins("LEFT JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.id = ? AND (jobs.company_id = ? OR applications.job_id IS NULL)", applicationID, companyID).
		Preload("Job").
		First(&application).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	// Store application details for logging before deletion
	applicantName := application.FullName
	jobTitle := "Unknown Job (Job Deleted)"
	if application.JobID != nil && application.Job.ID != uuid.Nil {
		jobTitle = application.Job.Title
	}
	applicationUUID, _ := uuid.Parse(applicationID)
	companyUUID, _ := uuid.Parse(companyID)
	adminIDVal, _ := c.Get("admin_id")
	adminIDStr, _ := adminIDVal.(string)
	adminUUID, _ := uuid.Parse(adminIDStr)

	// Delete the application
	if err := config.DB.Delete(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete application"})
		return
	}

	// Log application deletion
	services.LogActivity(
		&companyUUID,
		&adminUUID,
		"application_deleted",
		"application",
		&applicationUUID,
		"Application deleted: "+applicantName+" for job: "+jobTitle,
		map[string]interface{}{
			"applicant_name": applicantName,
			"applicant_email": application.Email,
			"job_title": jobTitle,
			"status": application.Status,
		},
	)

	c.JSON(http.StatusOK, gin.H{"message": "Application deleted successfully"})
}

// BulkDeleteApplications deletes multiple applications by status
func BulkDeleteApplications(c *gin.Context) {
	companyIDVal, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in token"})
		return
	}
	companyID, ok := companyIDVal.(string)
	if !ok || companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"` // pending, shortlisted, rejected
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	validStatuses := []string{"pending", "shortlisted", "rejected"}
	isValid := false
	for _, s := range validStatuses {
		if req.Status == s {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be: pending, shortlisted, or rejected"})
		return
	}

	// Get applications to delete
	var applications []models.Application
	err := config.DB.Table("applications").
		Select("applications.*").
		Joins("LEFT JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.status = ? AND (applications.company_id = ? OR jobs.company_id = ?)", req.Status, companyID, companyID).
		Preload("Job").
		Find(&applications).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
		return
	}

	if len(applications) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No applications found with the specified status",
			"deleted_count": 0,
		})
		return
	}

	// Get admin ID for logging
	adminIDVal, _ := c.Get("admin_id")
	adminIDStr, _ := adminIDVal.(string)
	adminUUID, _ := uuid.Parse(adminIDStr)
	companyUUID, _ := uuid.Parse(companyID)

	// Delete applications
	deletedCount := 0
	for _, app := range applications {
		if err := config.DB.Delete(&app).Error; err == nil {
			deletedCount++
			// Log each deletion
			appUUID, _ := uuid.Parse(app.ID.String())
			jobTitle := "Unknown Job (Job Deleted)"
			if app.JobID != nil && app.Job.ID != uuid.Nil {
				jobTitle = app.Job.Title
			}
			services.LogActivity(
				&companyUUID,
				&adminUUID,
				"application_deleted",
				"application",
				&appUUID,
				"Bulk deleted application: "+app.FullName+" for job: "+jobTitle,
				map[string]interface{}{
					"applicant_name": app.FullName,
					"applicant_email": app.Email,
					"job_title": jobTitle,
					"status": app.Status,
					"bulk_delete": true,
				},
			)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Deleted %d application(s) with status '%s'", deletedCount, req.Status),
		"deleted_count": deletedCount,
		"total_found": len(applications),
	})
}

