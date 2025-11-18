package controllers

import (
	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SearchCandidatesRequest defines the search criteria
type SearchCandidatesRequest struct {
	Query            string   `json:"query"`             // General text search
	Skills           []string `json:"skills"`            // Required skills
	MinExperience    *int     `json:"min_experience"`    // Minimum years of experience
	MaxExperience    *int     `json:"max_experience"`    // Maximum years of experience
	CurrentPosition  string   `json:"current_position"`  // Current position keyword
	Languages        []string `json:"languages"`         // Required languages
	HasPortfolio     *bool    `json:"has_portfolio"`     // Has portfolio URL
	HasLinkedIn      *bool    `json:"has_linkedin"`      // Has LinkedIn URL
	Status           string   `json:"status"`            // Application status filter
	Limit            int      `json:"limit"`             // Results limit
}

// CandidateSearchResult represents a search result
type CandidateSearchResult struct {
	Application    models.Application `json:"application"`
	MatchScore     int                `json:"match_score"`     // 0-100
	MatchedSkills  []string           `json:"matched_skills"`  // Skills found
	MatchedReasons []string           `json:"matched_reasons"` // Why it matched
}

// SearchCandidates searches through all CVs in the database
func SearchCandidates(c *gin.Context) {
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

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID format"})
		return
	}

	var req SearchCandidatesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default limit
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	// Get all applications for this company's jobs
	var applications []models.Application
	query := config.DB.Joins("JOIN jobs ON jobs.id = applications.job_id").
		Where("jobs.company_id = ?", companyID).
		Preload("Job")

	// Apply status filter if provided
	if req.Status != "" {
		query = query.Where("applications.status = ?", req.Status)
	}

	// Apply filters
	if req.HasPortfolio != nil && *req.HasPortfolio {
		query = query.Where("applications.portfolio_url IS NOT NULL AND applications.portfolio_url != ''")
	}

	if req.HasLinkedIn != nil && *req.HasLinkedIn {
		query = query.Where("applications.linkedin_url IS NOT NULL AND applications.linkedin_url != ''")
	}

	if req.MinExperience != nil {
		query = query.Where("applications.years_of_experience >= ?", *req.MinExperience)
	}

	if req.MaxExperience != nil {
		query = query.Where("applications.years_of_experience <= ?", *req.MaxExperience)
	}

	if req.CurrentPosition != "" {
		query = query.Where("LOWER(applications.current_position) LIKE ?", "%"+strings.ToLower(req.CurrentPosition)+"%")
	}

	// Execute query
	if err := query.Find(&applications).Error; err != nil {
		log.Printf("Error fetching applications: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch candidates"})
		return
	}

	// Search through CVs
	results := []CandidateSearchResult{}

	for _, app := range applications {
		matchScore := 0
		matchedSkills := []string{}

		// Get CV text (from parsed_cv_text or extract from URL)
		cvText := ""
		if app.ParsedCVText != nil && *app.ParsedCVText != "" {
			cvText = *app.ParsedCVText
		} else if app.ResumeURL != "" {
			// Extract text from URL if not parsed yet
			extractedText, err := services.ExtractTextFromURL(app.ResumeURL)
			if err == nil && len(extractedText) > 50 {
				cvText = extractedText
				// Store parsed text for future searches
				app.ParsedCVText = &extractedText
				config.DB.Model(&app).Update("parsed_cv_text", extractedText)
			}
		}

		if cvText == "" {
			// Skip if no CV text available
			continue
		}

		cvLower := strings.ToLower(cvText)
		reasons := []string{}

		// 1. General text query search
		hasQueryMatch := false
		if req.Query != "" {
			queryLower := strings.ToLower(req.Query)
			queryWords := strings.Fields(queryLower)
			matchedWords := 0

			for _, word := range queryWords {
				if len(word) > 2 && strings.Contains(cvLower, word) {
					matchedWords++
				}
			}

			if matchedWords > 0 {
				queryScore := (matchedWords * 100) / len(queryWords)
				if queryScore > 50 { // Only include if significant match
					matchScore += queryScore
					hasQueryMatch = true
					reasons = append(reasons, fmt.Sprintf("Matched %d/%d search terms", matchedWords, len(queryWords)))
				}
			}
		} else {
			// No query means we'll match based on other criteria
			hasQueryMatch = true
		}

		// 2. Skills search
		if len(req.Skills) > 0 {
			foundSkills := services.ExtractSkills(cvText, req.Skills)
			matchedSkills = foundSkills
			if len(foundSkills) > 0 {
				skillsScore := (len(foundSkills) * 100) / len(req.Skills)
				matchScore += skillsScore
				reasons = append(reasons, fmt.Sprintf("Found %d/%d required skills: %s", len(foundSkills), len(req.Skills), strings.Join(foundSkills, ", ")))
			}
		}

		// 3. Experience search
		if req.MinExperience != nil {
			if app.YearsOfExperience >= *req.MinExperience {
				matchScore += 20
				reasons = append(reasons, fmt.Sprintf("Has %d years of experience (required: %d+)", app.YearsOfExperience, *req.MinExperience))
			}
		}

		// 4. Languages search
		if len(req.Languages) > 0 {
			foundLanguages := services.ExtractLanguages(cvText, req.Languages)
			if len(foundLanguages) > 0 {
				langScore := (len(foundLanguages) * 100) / len(req.Languages)
				matchScore += langScore / 5 // 20% weight
				reasons = append(reasons, fmt.Sprintf("Found languages: %s", strings.Join(foundLanguages, ", ")))
			}
		}

		// 5. Current position search
		if req.CurrentPosition != "" {
			if strings.Contains(strings.ToLower(app.CurrentPosition), strings.ToLower(req.CurrentPosition)) {
				matchScore += 15
				reasons = append(reasons, fmt.Sprintf("Current position matches: %s", app.CurrentPosition))
			}
		}

		// Only include if there's a match
		// If query was provided, it must match. Otherwise, check other criteria
		hasOtherCriteria := len(req.Skills) > 0 || req.MinExperience != nil || req.MaxExperience != nil ||
			req.CurrentPosition != "" || len(req.Languages) > 0 || req.HasPortfolio != nil || req.HasLinkedIn != nil || req.Status != ""
		
		if hasQueryMatch {
			// If query matches (or no query provided), check if we have other criteria matches
			if hasOtherCriteria {
				// Need at least some match from other criteria
				if matchScore > 0 {
					// Normalize score to 0-100
					if matchScore > 100 {
						matchScore = 100
					}
					results = append(results, CandidateSearchResult{
						Application:    app,
						MatchScore:     matchScore,
						MatchedSkills:  matchedSkills,
						MatchedReasons: reasons,
					})
				}
			} else {
				// No other criteria, just query match (or no query at all) - show all
				if matchScore == 0 && req.Query == "" {
					matchScore = 50 // Default score if no criteria
				}
				// Normalize score to 0-100
				if matchScore > 100 {
					matchScore = 100
				}
				results = append(results, CandidateSearchResult{
					Application:    app,
					MatchScore:     matchScore,
					MatchedSkills:  matchedSkills,
					MatchedReasons: reasons,
				})
			}
		}
	}

	// Sort by match score (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].MatchScore < results[j].MatchScore {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Limit results
	if len(results) > req.Limit {
		results = results[:req.Limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"candidates": results,
		"count":      len(results),
		"total":      len(applications),
	})
}

// GetCandidateDetails returns detailed information about a candidate
func GetCandidateDetails(c *gin.Context) {
	candidateID := c.Param("id")
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

	var application models.Application
	err := config.DB.Joins("JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.id = ? AND jobs.company_id = ?", candidateID, companyIDStr).
		Preload("Job").
		First(&application).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Candidate not found"})
		return
	}

	// Get CV text if available
	cvText := ""
	if application.ParsedCVText != nil && *application.ParsedCVText != "" {
		cvText = *application.ParsedCVText
	} else if application.ResumeURL != "" {
		// Extract if not parsed
		extractedText, err := services.ExtractTextFromURL(application.ResumeURL)
		if err == nil {
			cvText = extractedText
			// Store for future
			application.ParsedCVText = &extractedText
			config.DB.Model(&application).Update("parsed_cv_text", extractedText)
		}
	}

	// Extract skills and experience from CV
	skills := []string{}
	experience := 0
	if cvText != "" {
		skills = services.ExtractSkills(cvText, []string{}) // Extract all skills
		experience = services.ExtractExperience(cvText)
	}

	c.JSON(http.StatusOK, gin.H{
		"candidate": application,
		"cv_text":   cvText,
		"skills":    skills,
		"experience": experience,
	})
}

