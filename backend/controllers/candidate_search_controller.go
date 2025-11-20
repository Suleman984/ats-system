package controllers

import (
	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"
	"fmt"
	"net/http"
	"regexp"
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
	fmt.Println("companyIDVal", companyIDVal)
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

	// Get all applications for this company, including those with deleted jobs (job_id IS NULL)
	// This allows Find Candidates to search through ALL applications, even if their jobs were deleted
	var applications []models.Application
	
	// Query 1: Applications with active jobs (uses index on jobs.company_id and applications.job_id)
	// This should always work regardless of company_id column existence
	var activeJobApps []models.Application
	err1 := config.DB.Table("applications").
		Select("applications.*").
		Joins("INNER JOIN jobs ON jobs.id = applications.job_id").
		Where("jobs.company_id = ?", companyID).
		Preload("Job").
		Find(&activeJobApps).Error
		
	// Debug: Check if ParsedCVText is loaded
	if len(activeJobApps) > 0 {
		fmt.Printf("DEBUG: First active app - Has ParsedCVText: %v, Length: %d\n",
			activeJobApps[0].ParsedCVText != nil,
			func() int {
				if activeJobApps[0].ParsedCVText != nil {
					return len(*activeJobApps[0].ParsedCVText)
				}
				return 0
			}())
	}
	
	if err1 != nil {
		fmt.Printf("ERROR: Failed to fetch active job applications: %v\n", err1)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch candidates", "details": err1.Error()})
		return
	}
	
	// Query 2: Applications with deleted jobs (job_id IS NULL)
	// Try with company_id first (if column exists and is populated)
	var deletedJobApps []models.Application
	err2 := config.DB.Table("applications").
		Select("applications.*").
		Where("applications.job_id IS NULL AND applications.company_id = ?", companyID).
		Preload("Job").
		Find(&deletedJobApps).Error
	
	// If query fails, it might mean:
	// 1. company_id column doesn't exist yet (migration not run)
	// 2. company_id is NULL for those applications
	// In either case, we can't reliably get deleted job applications, so we'll skip them
	if err2 != nil {
		fmt.Printf("INFO: Could not fetch deleted job applications (company_id column may not exist or be NULL): %v\n", err2)
		fmt.Printf("INFO: This is OK if you haven't run ADD_COMPANY_ID_TO_APPLICATIONS.sql migration yet\n")
		// Continue with just active job applications
		deletedJobApps = []models.Application{}
	}
	
	// Combine results
	applications = append(activeJobApps, deletedJobApps...)
	
	// Debug logging
	fmt.Printf("DEBUG: Found %d total applications (%d active jobs, %d deleted jobs) for company %s\n", 
		len(applications), len(activeJobApps), len(deletedJobApps), companyID.String())
	fmt.Printf("DEBUG: Search request - Query: '%s', Skills: %v, MinExp: %v, Languages: %v\n",
		req.Query, req.Skills, req.MinExperience, req.Languages)

	// Apply filters to the combined results (in-memory filtering for better performance)
	filteredApplications := []models.Application{}
	for _, app := range applications {
		fmt.Printf("DEBUG: Processing application %s - Email: %s, Has CV Text: %v\n",
			app.ID.String(), app.Email, app.ParsedCVText != nil && *app.ParsedCVText != "")
		// Status filter
		if req.Status != "" && app.Status != req.Status {
			continue
		}
		
		// Portfolio filter
		if req.HasPortfolio != nil && *req.HasPortfolio {
			if app.PortfolioURL == "" {
				continue
			}
		}
		
		// LinkedIn filter
		if req.HasLinkedIn != nil && *req.HasLinkedIn {
			if app.LinkedinURL == "" {
				continue
			}
		}
		
		// Experience filters
		if req.MinExperience != nil && app.YearsOfExperience < *req.MinExperience {
			continue
		}
		if req.MaxExperience != nil && app.YearsOfExperience > *req.MaxExperience {
			continue
		}
		
		// Current position filter
		if req.CurrentPosition != "" {
			if !strings.Contains(strings.ToLower(app.CurrentPosition), strings.ToLower(req.CurrentPosition)) {
				continue
			}
		}
		
		filteredApplications = append(filteredApplications, app)
	}
	
	applications = filteredApplications
	
	fmt.Printf("DEBUG: After filtering, %d applications remain\n", len(applications))

	// Search through CVs
	results := []CandidateSearchResult{}

	for _, app := range applications {
		fmt.Printf("DEBUG: Analyzing application %s - CV Text Length: %d\n",
			app.Email, func() int {
				if app.ParsedCVText != nil {
					return len(*app.ParsedCVText)
				}
				return 0
			}())
		matchScore := 0
		matchedSkills := []string{}

		// Get CV text (from parsed_cv_text or extract from URL)
		cvText := ""
		if app.ParsedCVText != nil && *app.ParsedCVText != "" {
			cvText = *app.ParsedCVText
			fmt.Printf("DEBUG: Using parsed CV text for %s (%d chars)\n", app.Email, len(cvText))
		} else if app.ResumeURL != "" {
			// Extract text from URL if not parsed yet
			fmt.Printf("DEBUG: No parsed CV text, extracting from URL for %s\n", app.Email)
			extractedText, err := services.ExtractTextFromURL(app.ResumeURL)
			if err == nil && len(extractedText) > 50 {
				cvText = extractedText
				// Store parsed text for future searches
				app.ParsedCVText = &extractedText
				config.DB.Model(&app).Update("parsed_cv_text", extractedText)
				fmt.Printf("DEBUG: Extracted and stored CV text for %s (%d chars)\n", app.Email, len(cvText))
			} else {
				fmt.Printf("DEBUG: Failed to extract CV text for %s: %v\n", app.Email, err)
			}
		}

		if cvText == "" {
			// If no CV text but we have other search criteria, still include the candidate
			// Only skip if we need CV text for the search
			if req.Query != "" || len(req.Skills) > 0 || len(req.Languages) > 0 {
				// Need CV text for these searches, skip this candidate
				fmt.Printf("DEBUG: Skipping %s - no CV text but search requires it\n", app.Email)
				continue
			}
			// Otherwise, continue with empty CV text (will match based on other fields)
			fmt.Printf("DEBUG: Continuing with %s - no CV text but search doesn't require it\n", app.Email)
		}

		cvLower := strings.ToLower(cvText)
		reasons := []string{}

		// 1. General text query search
		hasQueryMatch := false
		if req.Query != "" {
			queryLower := strings.ToLower(strings.TrimSpace(req.Query))
			queryWords := strings.Fields(queryLower)
			matchedWords := 0

			for _, word := range queryWords {
				// Match words longer than 2 characters (more reliable)
				// Use word boundaries for better matching
				if len(word) > 2 {
					// Check for exact word match or as part of a larger word
					wordPattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`)
					if wordPattern.MatchString(cvLower) {
						matchedWords++
					} else if strings.Contains(cvLower, word) {
						// Fallback: partial match
						matchedWords++
					}
				}
			}

			if matchedWords > 0 {
				queryScore := (matchedWords * 100) / len(queryWords)
				// Include if at least 1 word matches
				matchScore += queryScore
				hasQueryMatch = true
				reasons = append(reasons, fmt.Sprintf("Matched %d/%d search terms", matchedWords, len(queryWords)))
			}
		} else {
			// No query means we'll match based on other criteria
			hasQueryMatch = true
		}

		// 2. Skills search
		if len(req.Skills) > 0 {
			foundSkills := services.ExtractSkills(cvText, req.Skills)
			matchedSkills = foundSkills
			fmt.Printf("DEBUG: Skills search for %s - Required: %v, Found: %v\n", app.Email, req.Skills, foundSkills)
			if len(foundSkills) > 0 {
				skillsScore := (len(foundSkills) * 100) / len(req.Skills)
				matchScore += skillsScore
				reasons = append(reasons, fmt.Sprintf("Found %d/%d required skills: %s", len(foundSkills), len(req.Skills), strings.Join(foundSkills, ", ")))
			} else {
				fmt.Printf("DEBUG: No skills matched for %s\n", app.Email)
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
		
		// Include candidate if:
		// 1. Query matches (or no query provided) AND
		// 2. Either no other criteria OR at least one criterion matches
		shouldInclude := false
		
		if hasQueryMatch {
			if hasOtherCriteria {
				// Have other criteria - need at least some match
				if matchScore > 0 {
					shouldInclude = true
				}
			} else {
				// No other criteria - include if query matched (or no query)
				shouldInclude = true
			}
		}
		
		if shouldInclude {
			// Set default score if no criteria matched
			if matchScore == 0 && req.Query == "" && !hasOtherCriteria {
				matchScore = 50 // Default score if no criteria at all
			}
			
			// Normalize score to 0-100
			if matchScore > 100 {
				matchScore = 100
			}
			
			fmt.Printf("DEBUG: Including candidate %s with score %d, reasons: %v\n", app.Email, matchScore, reasons)
			results = append(results, CandidateSearchResult{
				Application:    app,
				MatchScore:     matchScore,
				MatchedSkills:  matchedSkills,
				MatchedReasons: reasons,
			})
		} else {
			fmt.Printf("DEBUG: Excluding candidate %s - hasQueryMatch: %v, hasOtherCriteria: %v, matchScore: %d\n",
				app.Email, hasQueryMatch, hasOtherCriteria, matchScore)
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

	// Debug logging
	fmt.Printf("DEBUG: Search completed. Found %d matching candidates out of %d total applications\n", 
		len(results), len(applications))
	fmt.Printf("DEBUG: Active job apps: %d, Deleted job apps: %d\n", len(activeJobApps), len(deletedJobApps))

	c.JSON(http.StatusOK, gin.H{
		"candidates": results,
		"count":      len(results),
		"total":      len(applications),
		"debug": gin.H{
			"active_job_apps": len(activeJobApps),
			"deleted_job_apps": len(deletedJobApps),
			"total_applications_before_filter": len(activeJobApps) + len(deletedJobApps),
			"total_applications_after_filter": len(applications),
			"matching_candidates": len(results),
		},
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
	// Allow viewing candidates even if their job was deleted (for Find Candidates)
	// OPTIMIZED: Try active job first (most common case, uses index), then deleted job
	err := config.DB.Table("applications").
		Select("applications.*").
		Joins("INNER JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.id = ? AND jobs.company_id = ?", candidateID, companyIDStr).
		Preload("Job").
		First(&application).Error
	
	// If not found, try deleted job applications
	if err != nil {
		err = config.DB.Table("applications").
			Select("applications.*").
			Where("applications.id = ? AND applications.job_id IS NULL AND applications.company_id = ?", candidateID, companyIDStr).
			Preload("Job").
			First(&application).Error
	}

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

