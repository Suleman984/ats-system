package controllers

import (
	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ManualCandidateRequest for adding candidates manually
type ManualCandidateRequest struct {
	JobID              string `json:"job_id" binding:"required"`
	FullName           string `json:"full_name" binding:"required"`
	Email              string `json:"email" binding:"required,email"`
	Phone              string `json:"phone"`
	ResumeURL          string `json:"resume_url" binding:"required"`
	CoverLetter        string `json:"cover_letter"`
	YearsOfExperience  int    `json:"years_of_experience"`
	CurrentPosition    string `json:"current_position"`
	LinkedinURL        string `json:"linkedin_url"`
	PortfolioURL       string `json:"portfolio_url"`
	Status             string `json:"status"` // Can be "pending", "shortlisted", etc.
	Notes              string `json:"notes"` // Admin notes about why this candidate was added manually
}

// AddManualCandidate allows admins to manually add candidates that AI might have missed
func AddManualCandidate(c *gin.Context) {
	companyIDVal, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in token"})
		return
	}

	companyIDStr, ok := companyIDVal.(string)
	if !ok || companyIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var req ManualCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify job belongs to company
	jobID, err := uuid.Parse(req.JobID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	var job models.Job
	if err := config.DB.Where("id = ? AND company_id = ?", jobID, companyIDStr).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found or doesn't belong to your company"})
		return
	}

	// Check if candidate already applied for this job
	var existingApp models.Application
	if err := config.DB.Where("job_id = ? AND email = ?", jobID, req.Email).First(&existingApp).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Candidate already applied for this job",
			"application_id": existingApp.ID,
		})
		return
	}

	// Get admin ID for logging
	adminIDVal, _ := c.Get("admin_id")
	adminIDStr, _ := adminIDVal.(string)
	adminID, _ := uuid.Parse(adminIDStr)
	
	// Parse company ID
	companyUUID, _ := uuid.Parse(companyIDStr)

	// Create application
	application := models.Application{
		JobID:             &jobID,
		CompanyID:         companyUUID,
		FullName:          req.FullName,
		Email:             req.Email,
		Phone:              req.Phone,
		ResumeURL:          req.ResumeURL,
		CoverLetter:       req.CoverLetter,
		YearsOfExperience: req.YearsOfExperience,
		CurrentPosition:   req.CurrentPosition,
		LinkedinURL:       req.LinkedinURL,
		PortfolioURL:      req.PortfolioURL,
		Status:             req.Status,
		AppliedAt:         time.Now(),
	}

	if application.Status == "" {
		application.Status = "pending"
	}

	// Save application
	if err := config.DB.Create(&application).Error; err != nil {
		log.Printf("Failed to create manual candidate: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add candidate"})
		return
	}

	// Parse and store CV text (async)
	go func() {
		if application.ResumeURL != "" {
			log.Printf("Parsing CV text for manually added candidate %s (URL: %s)", application.ID.String(), application.ResumeURL)
			cvText, err := services.ExtractTextFromURL(application.ResumeURL)
			if err != nil {
				log.Printf("ERROR: Failed to parse CV text for %s (Email: %s, URL: %s): %v", 
					application.FullName, application.Email, application.ResumeURL, err)
				return
			}
			
			if len(cvText) < 50 {
				log.Printf("WARNING: CV text too short for %s (%d characters). URL: %s", 
					application.FullName, len(cvText), application.ResumeURL)
				return
			}
			
			// Validate UTF-8 before saving
			if !utf8.ValidString(cvText) {
				log.Printf("ERROR: Extracted CV text is not valid UTF-8 for %s. Skipping save.", 
					application.FullName)
				return
			}
			
			// Update using the application ID (more reliable than using the model)
			updateErr := config.DB.Model(&models.Application{}).
				Where("id = ?", application.ID).
				Update("parsed_cv_text", cvText).Error
			
			if updateErr != nil {
				log.Printf("ERROR: Failed to save parsed CV text to database for %s: %v", 
					application.FullName, updateErr)
			} else {
				log.Printf("SUCCESS: CV text parsed and stored for %s (%d characters)", 
					application.FullName, len(cvText))
			}
		} else {
			log.Printf("WARNING: No ResumeURL provided for manually added application %s", application.ID.String())
		}
	}()

	// Auto-analyze CV if job has criteria (async)
	go func() {
		log.Printf("Auto-analyzing CV for manually added candidate %s", application.ID.String())
		
		var criteria services.Criteria
		
		// Try to use job's shortlist criteria if available
		if job.ShortlistCriteria != nil && *job.ShortlistCriteria != "" {
			if err := json.Unmarshal([]byte(*job.ShortlistCriteria), &criteria); err == nil {
				criteria.JobDescription = job.Description
				criteria.JobRequirements = job.Requirements
			}
		} else {
			// Use default criteria based on job description
			criteria = services.Criteria{
				RequiredSkills:      []string{},
				MinExperience:       0,
				RequiredLanguages:   []string{},
				MatchJobDescription: true,
				JobDescription:      job.Description,
				JobRequirements:     job.Requirements,
			}
		}
		
		// Analyze CV
		analysisResult, err := services.MatchCVFromURL(application.ResumeURL, criteria, job.Title)
		if err == nil {
			application.Score = analysisResult.MatchScore
			analysisJSON, _ := json.Marshal(analysisResult)
			analysisJSONStr := string(analysisJSON)
			application.AnalysisResult = &analysisJSONStr
			config.DB.Save(&application)
			log.Printf("CV analyzed for manually added candidate %s: Score=%d%%", application.FullName, analysisResult.MatchScore)
		} else {
			log.Printf("Failed to analyze CV for manually added candidate %s: %v", application.Email, err)
		}
	}()

	// Log manual candidate addition
	services.LogActivity(
		&companyUUID,
		&adminID,
		"manual_candidate_added",
		"application",
		&application.ID,
		"Manually added candidate: "+req.FullName+" for job: "+job.Title,
		map[string]interface{}{
			"candidate_name": req.FullName,
			"candidate_email": req.Email,
			"job_title": job.Title,
			"notes": req.Notes,
		},
	)

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Candidate added successfully",
		"application": application,
	})
}

