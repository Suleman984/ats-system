package controllers

import (
	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AIShortlistRequest defines the criteria for AI shortlisting
type AIShortlistRequest struct {
	ApplicationID    string   `json:"application_id" binding:"required"`
	RequiredSkills   []string `json:"required_skills"`
	MinExperience    int      `json:"min_experience"`
	RequiredLanguages []string `json:"required_languages"`
	MatchJobDescription bool   `json:"match_job_description"`
}

// AIShortlistApplication analyzes and shortlists an application using AI
func AIShortlistApplication(c *gin.Context) {
	var req AIShortlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	// Get application
	var application models.Application
	if err := config.DB.Where("id = ?", req.ApplicationID).
		Preload("Job").
		First(&application).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	// Verify job belongs to company
	if application.Job.CompanyID.String() != companyID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Build criteria
	criteria := services.Criteria{
		RequiredSkills:      req.RequiredSkills,
		MinExperience:       req.MinExperience,
		RequiredLanguages:   req.RequiredLanguages,
		MatchJobDescription: req.MatchJobDescription,
		JobDescription:      application.Job.Description,
		JobRequirements:     application.Job.Requirements,
	}

	// If no specific criteria provided, use job's criteria
	if len(req.RequiredSkills) == 0 && req.MinExperience == 0 && len(req.RequiredLanguages) == 0 {
		// Try to parse job's shortlist criteria
		if application.Job.ShortlistCriteria != nil {
			var jobCriteria services.Criteria
			if err := json.Unmarshal([]byte(*application.Job.ShortlistCriteria), &jobCriteria); err == nil {
				criteria = jobCriteria
				criteria.JobDescription = application.Job.Description
				criteria.JobRequirements = application.Job.Requirements
			}
		}
	}

	// Analyze CV with local matching algorithm
	log.Printf("Analyzing CV for application %s using local matching", req.ApplicationID)
	analysisResult, err := services.MatchCVFromURL(application.ResumeURL, criteria, application.Job.Title)
	if err != nil {
		log.Printf("CV matching failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to analyze CV",
			"details": err.Error(),
		})
		return
	}

	// Update application with score and analysis result
	application.Score = analysisResult.MatchScore
	
	// Store analysis result as JSON
	analysisJSON, _ := json.Marshal(analysisResult)
	analysisJSONStr := string(analysisJSON)
	application.AnalysisResult = &analysisJSONStr
	
	log.Printf("CV Matching Result for %s: Score=%d%%, Skills=%v, Experience=%d years",
		application.FullName, analysisResult.MatchScore, analysisResult.Skills, analysisResult.Experience)

	// Note: We only calculate and store the score
	// Admin decides whether to shortlist based on the score

	// Save updated application
	if err := config.DB.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "CV analyzed successfully. Review the score and shortlist manually if needed.",
		"application": application,
		"analysis":    analysisResult,
	})
}

// BatchAIShortlist analyzes multiple applications
func BatchAIShortlist(c *gin.Context) {
	var req struct {
		JobID            string   `json:"job_id" binding:"required"`
		RequiredSkills   []string `json:"required_skills"`
		MinExperience    int      `json:"min_experience"`
		RequiredLanguages []string `json:"required_languages"`
		MatchJobDescription bool   `json:"match_job_description"`
		Threshold        int      `json:"threshold"` // Auto-shortlist threshold (default 70)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
	threshold := req.Threshold
	if threshold == 0 {
		threshold = 70 // Default threshold
	}

	// Get job
	var job models.Job
	if err := config.DB.Where("id = ? AND company_id = ?", req.JobID, companyID).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	// Get pending applications for this job
	var applications []models.Application
	if err := config.DB.Where("job_id = ? AND status = ?", req.JobID, "pending").Find(&applications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
		return
	}

	// Build criteria
	criteria := services.Criteria{
		RequiredSkills:      req.RequiredSkills,
		MinExperience:       req.MinExperience,
		RequiredLanguages:   req.RequiredLanguages,
		MatchJobDescription: req.MatchJobDescription,
		JobDescription:      job.Description,
		JobRequirements:     job.Requirements,
	}

	// If no criteria provided, use job's criteria
	if len(req.RequiredSkills) == 0 && req.MinExperience == 0 && len(req.RequiredLanguages) == 0 {
		if job.ShortlistCriteria != nil {
			var jobCriteria services.Criteria
			if err := json.Unmarshal([]byte(*job.ShortlistCriteria), &jobCriteria); err == nil {
				criteria = jobCriteria
				criteria.JobDescription = job.Description
				criteria.JobRequirements = job.Requirements
			}
		}
	}

	results := make([]map[string]interface{}, 0)

		// Analyze each application
	for _, app := range applications {
		log.Printf("Matching CV for %s (Application ID: %s)", app.FullName, app.ID.String())
		
		analysisResult, err := services.MatchCVFromURL(app.ResumeURL, criteria, job.Title)
		if err != nil {
			log.Printf("Failed to match CV for %s: %v", app.FullName, err)
			results = append(results, map[string]interface{}{
				"application_id": app.ID.String(),
				"candidate_name": app.FullName,
				"error":          err.Error(),
			})
			continue
		}

		// Update application score and analysis
		app.Score = analysisResult.MatchScore
		analysisJSON, _ := json.Marshal(analysisResult)
		analysisJSONStr := string(analysisJSON)
		app.AnalysisResult = &analysisJSONStr
		
		// Note: We only calculate and store the score
		// Admin decides whether to shortlist based on the score
		// No auto-shortlisting - admin has full control

		// Save application
		config.DB.Save(&app)

		results = append(results, map[string]interface{}{
			"application_id": app.ID.String(),
			"candidate_name": app.FullName,
			"match_score":    analysisResult.MatchScore,
			"analysis":       analysisResult,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        fmt.Sprintf("Analyzed %d applications. Review scores and shortlist manually.", len(applications)),
		"total_analyzed": len(applications),
		"results":        results,
	})
}

