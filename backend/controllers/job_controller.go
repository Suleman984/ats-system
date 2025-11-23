package controllers

import (
	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Helper function to get company_id from context
func getCompanyID(c *gin.Context) (string, error) {
	companyIDVal, exists := c.Get("company_id")
	if !exists {
		return "", gin.Error{Err: nil, Type: gin.ErrorTypePublic, Meta: "Company ID not found in token"}
	}
	companyID, ok := companyIDVal.(string)
	if !ok || companyID == "" {
		return "", gin.Error{Err: nil, Type: gin.ErrorTypePublic, Meta: "Invalid company ID"}
	}
	return companyID, nil
}

// CreateJob creates a new job posting
func CreateJob(c *gin.Context) {
	var jobRequest struct {
		Title            string `json:"title" binding:"required"`
		Description      string `json:"description" binding:"required"`
		Requirements     string `json:"requirements"`
		Location         string `json:"location"`
		JobType          string `json:"job_type"`
		SalaryRange      string `json:"salary_range"`
		Deadline         string `json:"deadline" binding:"required"`
		Status           string `json:"status"`
		AutoShortlist    bool   `json:"auto_shortlist"`
		ShortlistCriteria string `json:"shortlist_criteria"`
	}

	if err := c.ShouldBindJSON(&jobRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get company ID from JWT token (set by auth middleware)
	companyIDVal, exists := c.Get("company_id")
	if !exists {
		log.Printf("CreateJob: ERROR - company_id not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in token"})
		return
	}
	
	companyIDStr, ok := companyIDVal.(string)
	if !ok || companyIDStr == "" {
		log.Printf("CreateJob: ERROR - company_id is invalid, type: %T, value: '%v'", companyIDVal, companyIDVal)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Company ID is invalid",
			"details": "Your authentication token does not contain a valid company ID. Please log out and log in again.",
		})
		return
	}
	
	log.Printf("CreateJob: Extracted company_id: '%s'", companyIDStr)
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		log.Printf("CreateJob: ERROR - Failed to parse company_id '%s': %v", companyIDStr, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid company ID format",
			"details": "The company ID in your token is not in a valid format.",
		})
		return
	}

	// Parse deadline date string (format: "YYYY-MM-DD")
	deadlineStr := strings.TrimSpace(jobRequest.Deadline)
	deadlineTime, err := time.Parse("2006-01-02", deadlineStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid deadline format. Use YYYY-MM-DD",
			"details": err.Error(),
			"received": deadlineStr,
		})
		return
	}
	
	// Create DateOnly type
	deadline := models.DateOnly{
		Time: time.Date(deadlineTime.Year(), deadlineTime.Month(), deadlineTime.Day(), 0, 0, 0, 0, time.UTC),
	}

	// Create job model
	job := models.Job{
		CompanyID:        companyID,
		Title:            jobRequest.Title,
		Description:      jobRequest.Description,
		Requirements:     jobRequest.Requirements,
		Location:         jobRequest.Location,
		JobType:          jobRequest.JobType,
		SalaryRange:      jobRequest.SalaryRange,
		Deadline:         deadline,
		Status:           jobRequest.Status,
		AutoShortlist:    jobRequest.AutoShortlist,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if job.Status == "" {
		job.Status = "open"
	}
	if !job.AutoShortlist {
		job.AutoShortlist = true
	}
	
	// Handle shortlist_criteria - only set if not empty, otherwise leave as NULL
	if jobRequest.ShortlistCriteria != "" {
		job.ShortlistCriteria = &jobRequest.ShortlistCriteria
	}

	// Save to database
	if err := config.DB.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create job",
			"details": err.Error(),
		})
		return
	}

	// Get admin ID from context
	adminIDVal, _ := c.Get("admin_id")
	adminIDStr, _ := adminIDVal.(string)
	adminID, _ := uuid.Parse(adminIDStr)

	// Log job creation
	services.LogJobCreated(companyID, adminID, job.ID, job.Title)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Job created successfully",
		"job":     job,
	})
}

// GetJobs returns all jobs for a company
func GetJobs(c *gin.Context) {
	companyID, err := getCompanyID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in token"})
		return
	}
	var jobs []models.Job

	// Auto-close expired jobs for this company before fetching
	// A job is considered expired if its deadline date is before today's date.
	today := time.Now().UTC().Format("2006-01-02")
	if err := config.DB.
		Model(&models.Job{}).
		Where("company_id = ? AND status = ? AND deadline < ?", companyID, "open", today).
		Update("status", "closed").Error; err != nil {
		log.Printf("GetJobs: failed to auto-close expired jobs for company %s: %v", companyID, err)
	}

	query := config.DB.Where("company_id = ?", companyID)

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch jobs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

// GetJob returns a single job by ID
func GetJob(c *gin.Context) {
	jobID := c.Param("id")
	companyID, err := getCompanyID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in token"})
		return
	}

	var job models.Job
	if err := config.DB.Where("id = ? AND company_id = ?", jobID, companyID).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"job": job})
}

// GetPublicJobs returns open jobs for public (no auth needed)
func GetPublicJobs(c *gin.Context) {
	companyID := c.Param("companyId")
	var jobs []models.Job

	now := time.Now()
	err := config.DB.Where("company_id = ? AND status = ? AND deadline > ?",
		companyID, "open", now.Format("2006-01-02")).Find(&jobs).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch jobs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

// UpdateJob updates an existing job
func UpdateJob(c *gin.Context) {
	jobID := c.Param("id")
	companyID, err := getCompanyID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in token"})
		return
	}

	var job models.Job
	if err := config.DB.Where("id = ? AND company_id = ?", jobID, companyID).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	// Store old values for logging
	oldStatus := job.Status
	oldTitle := job.Title
	jobUUID, _ := uuid.Parse(jobID)
	companyUUID, _ := uuid.Parse(companyID)
	adminIDVal, _ := c.Get("admin_id")
	adminIDStr, _ := adminIDVal.(string)
	adminUUID, _ := uuid.Parse(adminIDStr)

	var jobRequest struct {
		Title            string `json:"title"`
		Description      string `json:"description"`
		Requirements     string `json:"requirements"`
		Location         string `json:"location"`
		JobType          string `json:"job_type"`
		SalaryRange      string `json:"salary_range"`
		Deadline         string `json:"deadline"`
		Status           string `json:"status"`
		AutoShortlist    bool   `json:"auto_shortlist"`
		ShortlistCriteria string `json:"shortlist_criteria"`
	}

	if err := c.ShouldBindJSON(&jobRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if jobRequest.Title != "" {
		job.Title = jobRequest.Title
	}
	if jobRequest.Description != "" {
		job.Description = jobRequest.Description
	}
	if jobRequest.Requirements != "" {
		job.Requirements = jobRequest.Requirements
	}
	if jobRequest.Location != "" {
		job.Location = jobRequest.Location
	}
	if jobRequest.JobType != "" {
		job.JobType = jobRequest.JobType
	}
	if jobRequest.SalaryRange != "" {
		job.SalaryRange = jobRequest.SalaryRange
	}
	if jobRequest.Deadline != "" {
		deadlineStr := strings.TrimSpace(jobRequest.Deadline)
		deadlineTime, err := time.Parse("2006-01-02", deadlineStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid deadline format. Use YYYY-MM-DD",
				"details": err.Error(),
			})
			return
		}
		// Create DateOnly type
		job.Deadline = models.DateOnly{
			Time: time.Date(deadlineTime.Year(), deadlineTime.Month(), deadlineTime.Day(), 0, 0, 0, 0, time.UTC),
		}
	}
	if jobRequest.Status != "" {
		job.Status = jobRequest.Status
	}
	job.AutoShortlist = jobRequest.AutoShortlist
	if jobRequest.ShortlistCriteria != "" {
		job.ShortlistCriteria = &jobRequest.ShortlistCriteria
	} else {
		job.ShortlistCriteria = nil
	}

	job.UpdatedAt = time.Now()

	if err := config.DB.Save(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job"})
		return
	}

	// Log job update - check if status changed or other fields
	changes := make(map[string]interface{})
	if jobRequest.Status != "" && jobRequest.Status != oldStatus {
		services.LogJobStatusChanged(companyUUID, adminUUID, jobUUID, job.Title, oldStatus, job.Status)
	} else {
		// Log general update
		if jobRequest.Title != "" && jobRequest.Title != oldTitle {
			changes["title"] = map[string]interface{}{"old": oldTitle, "new": jobRequest.Title}
		}
		if len(changes) > 0 {
			services.LogJobUpdated(companyUUID, adminUUID, jobUUID, job.Title, changes)
		} else {
			services.LogJobUpdated(companyUUID, adminUUID, jobUUID, job.Title, map[string]interface{}{"updated": true})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Job updated successfully",
		"job":     job,
	})
}

// DeleteJob deletes a job
func DeleteJob(c *gin.Context) {
	jobID := c.Param("id")
	companyID, err := getCompanyID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in token"})
		return
	}

	var job models.Job
	if err := config.DB.Where("id = ? AND company_id = ?", jobID, companyID).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	// Store job details for logging before deletion
	jobTitle := job.Title
	jobUUID, _ := uuid.Parse(jobID)
	companyUUID, _ := uuid.Parse(companyID)
	adminIDVal, _ := c.Get("admin_id")
	adminIDStr, _ := adminIDVal.(string)
	adminUUID, _ := uuid.Parse(adminIDStr)

	if err := config.DB.Delete(&job).Error; err != nil {
		log.Printf("DeleteJob ERROR: Failed to delete job %s: %v", jobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete job",
			"details": err.Error(),
		})
		return
	}

	// Log job deletion (async, don't fail if logging fails)
	services.LogJobDeleted(companyUUID, adminUUID, jobUUID, jobTitle)

	c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}

