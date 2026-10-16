package controllers

import (
	"fmt"
	"log"
	"net/http"

	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Scoring is local keyword matching; the "AI" route names are kept for API compatibility.
type scoringCriteria struct {
	RequiredSkills      []string `json:"required_skills"`
	MinExperience       int      `json:"min_experience" binding:"min=0"`
	RequiredLanguages   []string `json:"required_languages"`
	MatchJobDescription bool     `json:"match_job_description"`
}

func (s scoringCriteria) forJob(job *models.Job) services.Criteria {
	if len(s.RequiredSkills) == 0 && s.MinExperience == 0 && len(s.RequiredLanguages) == 0 {
		return services.CriteriaForJob(job)
	}
	return services.Criteria{
		RequiredSkills:      s.RequiredSkills,
		MinExperience:       s.MinExperience,
		RequiredLanguages:   s.RequiredLanguages,
		MatchJobDescription: s.MatchJobDescription,
		JobDescription:      job.Description,
		JobRequirements:     job.Requirements,
	}
}

// scoreApplication reuses stored CV text when available to avoid re-downloading.
func scoreApplication(app *models.Application, criteria services.Criteria) (*services.MatchResult, error) {
	text := ""
	if app.ParsedCVText != nil {
		text = *app.ParsedCVText
	}
	if text == "" {
		var err error
		if text, err = services.ExtractTextFromURL(app.ResumeURL); err != nil {
			return nil, err
		}
	}
	result := services.MatchCV(text, criteria)
	if err := services.SaveAnalysis(app.ID, text, result); err != nil {
		return nil, err
	}
	return result, nil
}

func AIShortlistApplication(c *gin.Context) {
	var req struct {
		ApplicationID uuid.UUID `json:"application_id" binding:"required"`
		scoringCriteria
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	app, ok := findCompanyApplication(c, req.ApplicationID)
	if !ok {
		return
	}
	if app.Job == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The job for this application was deleted"})
		return
	}

	result, err := scoreApplication(app, req.forJob(app.Job))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to analyze CV", "details": err.Error()})
		return
	}
	app.Score = result.MatchScore

	c.JSON(http.StatusOK, gin.H{
		"message":     "CV analyzed. Review the score and shortlist manually if needed.",
		"application": app,
		"analysis":    result,
	})
}

func BatchAIShortlist(c *gin.Context) {
	var req struct {
		JobID uuid.UUID `json:"job_id" binding:"required"`
		scoringCriteria
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	job, ok := loadCompanyJob(c, req.JobID)
	if !ok {
		return
	}

	var applications []models.Application
	if err := config.DB.Where("job_id = ? AND status = ?", job.ID, models.StatusPending).Find(&applications).Error; err != nil {
		serverError(c, "Failed to fetch applications")
		return
	}

	criteria := req.forJob(job)
	results := make([]gin.H, 0, len(applications))
	for i := range applications {
		app := &applications[i]
		entry := gin.H{"application_id": app.ID, "candidate_name": app.FullName}
		if result, err := scoreApplication(app, criteria); err != nil {
			log.Printf("score application %s: %v", app.ID, err)
			entry["error"] = err.Error()
		} else {
			entry["match_score"], entry["analysis"] = result.MatchScore, result
		}
		results = append(results, entry)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        fmt.Sprintf("Analyzed %d applications. Review scores and shortlist manually.", len(applications)),
		"total_analyzed": len(applications),
		"results":        results,
	})
}
