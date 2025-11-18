package controllers

import (
	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"
	"encoding/json"
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
	if err := config.DB.First(&job, "id = ?", application.JobID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
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

	// Automatically analyze CV and calculate score if job has criteria (async)
	// Note: We only calculate score, admin decides whether to shortlist
	go func() {
		if job.ShortlistCriteria != nil && *job.ShortlistCriteria != "" {
			log.Printf("Auto-analyzing CV for application %s", application.ID.String())
			
			var criteria services.Criteria
			if err := json.Unmarshal([]byte(*job.ShortlistCriteria), &criteria); err == nil {
				criteria.JobDescription = job.Description
				criteria.JobRequirements = job.Requirements
				
				// Analyze CV
				analysisResult, err := services.MatchCVFromURL(application.ResumeURL, criteria, job.Title)
				if err != nil {
					log.Printf("Failed to auto-analyze CV for %s: %v", application.Email, err)
					return
				}
				
				// Update application with score (but don't auto-shortlist)
				application.Score = analysisResult.MatchScore
				analysisJSON, _ := json.Marshal(analysisResult)
				analysisJSONStr := string(analysisJSON)
				application.AnalysisResult = &analysisJSONStr
				
				// Save updated application with score
				config.DB.Save(&application)
				log.Printf("CV analyzed for %s: Score=%d%% (Admin can review and shortlist manually)", application.FullName, analysisResult.MatchScore)
			}
		}
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

	// Join with jobs table to filter by company
	query := config.DB.Joins("JOIN jobs ON jobs.id = applications.job_id").
		Where("jobs.company_id = ?", companyID).
		Preload("Job")

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
	// Verify application belongs to company and preload job
	err := config.DB.Joins("JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.id = ? AND jobs.company_id = ?", applicationID, companyID).
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
	jobTitle := application.Job.Title

	services.LogApplicationStatusChanged(companyUUID, adminUUID, applicationUUID, application.FullName, jobTitle, oldStatus, "shortlisted")

	// Send shortlist email (async with error logging)
	go func() {
		log.Printf("Sending shortlist email to %s for application %s", application.Email, applicationID)
		if err := services.SendShortlistEmail(application.Email, application.FullName); err != nil {
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
	err := config.DB.Joins("JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.id = ? AND jobs.company_id = ?", applicationID, companyID).
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
	jobTitle := application.Job.Title

	services.LogApplicationStatusChanged(companyUUID, adminUUID, applicationUUID, application.FullName, jobTitle, oldStatus, "rejected")

	// Send rejection email (async with error logging)
	go func() {
		log.Printf("Sending rejection email to %s", application.Email)
		if err := services.SendRejectionEmail(application.Email, application.FullName); err != nil {
			log.Printf("ERROR: Failed to send rejection email to %s: %v", application.Email, err)
		} else {
			log.Printf("SUCCESS: Rejection email sent to %s", application.Email)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Application rejected",
	})
}

