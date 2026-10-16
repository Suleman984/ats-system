package services

import (
	"encoding/json"
	"log"

	"ats-backend/config"
	"ats-backend/models"

	"github.com/google/uuid"
)

// CriteriaForJob uses the job's saved shortlist criteria, falling back to description matching.
func CriteriaForJob(job *models.Job) Criteria {
	criteria := Criteria{MatchJobDescription: true}
	if job.ShortlistCriteria != nil && *job.ShortlistCriteria != "" {
		if err := json.Unmarshal([]byte(*job.ShortlistCriteria), &criteria); err != nil {
			log.Printf("job %s: invalid shortlist criteria, using defaults: %v", job.ID, err)
			criteria = Criteria{MatchJobDescription: true}
		}
	}
	criteria.JobDescription = job.Description
	criteria.JobRequirements = job.Requirements
	return criteria
}

// SaveAnalysis stores the parsed text and score for an application.
func SaveAnalysis(applicationID uuid.UUID, cvText string, result *MatchResult) error {
	updates := map[string]any{"parsed_cv_text": cvText}
	if result != nil {
		raw, err := json.Marshal(result)
		if err != nil {
			return err
		}
		updates["score"] = result.MatchScore
		updates["analysis_result"] = string(raw)
	}
	return config.DB.Model(&models.Application{}).Where("id = ?", applicationID).Updates(updates).Error
}

// ProcessApplicationCV downloads the CV once, stores its text and match score.
func ProcessApplicationCV(applicationID uuid.UUID, resumeURL string, job *models.Job) {
	cvText, err := ExtractTextFromURL(resumeURL)
	if err != nil {
		log.Printf("application %s: CV extraction failed: %v", applicationID, err)
		return
	}
	var result *MatchResult
	if job != nil {
		result = MatchCV(cvText, CriteriaForJob(job))
	}
	if err := SaveAnalysis(applicationID, cvText, result); err != nil {
		log.Printf("application %s: saving CV analysis failed: %v", applicationID, err)
	}
}
