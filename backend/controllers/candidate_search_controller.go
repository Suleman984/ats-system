package controllers

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
)

type searchCandidatesRequest struct {
	Query           string   `json:"query" binding:"max=500"`
	Skills          []string `json:"skills" binding:"max=50"`
	MinExperience   *int     `json:"min_experience"`
	MaxExperience   *int     `json:"max_experience"`
	CurrentPosition string   `json:"current_position" binding:"max=255"`
	Languages       []string `json:"languages" binding:"max=20"`
	HasPortfolio    *bool    `json:"has_portfolio"`
	HasLinkedIn     *bool    `json:"has_linkedin"`
	Status          string   `json:"status"`
	Limit           int      `json:"limit"`
}

func (r *searchCandidatesRequest) needsCVText() bool {
	return r.Query != "" || len(r.Skills) > 0 || len(r.Languages) > 0
}

// softCriteria are scored rather than filtered in SQL, so a candidate must match at least one.
func (r *searchCandidatesRequest) softCriteria() bool {
	return len(r.Skills) > 0 || len(r.Languages) > 0
}

type candidateSearchResult struct {
	Application    models.Application `json:"application"`
	MatchScore     int                `json:"match_score"`
	MatchedSkills  []string           `json:"matched_skills"`
	MatchedReasons []string           `json:"matched_reasons"`
}

// SearchCandidates filters in SQL, then ranks the remaining CVs in memory.
// Candidates whose CV hasn't been parsed yet are skipped for text queries; use reparse to backfill.
func SearchCandidates(c *gin.Context) {
	var req searchCandidatesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	query := config.DB.Preload("Job").Where("company_id = ?", companyID(c))
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.HasPortfolio != nil && *req.HasPortfolio {
		query = query.Where("(coalesce(portfolio_url, '') <> '' OR portfolio_path IS NOT NULL)")
	}
	if req.HasLinkedIn != nil && *req.HasLinkedIn {
		query = query.Where("coalesce(linkedin_url, '') <> ''")
	}
	if req.MinExperience != nil {
		query = query.Where("years_of_experience >= ?", *req.MinExperience)
	}
	if req.MaxExperience != nil {
		query = query.Where("years_of_experience <= ?", *req.MaxExperience)
	}
	if req.CurrentPosition != "" {
		query = query.Where("current_position ILIKE ?", "%"+escapeLike(req.CurrentPosition)+"%")
	}
	if req.needsCVText() {
		query = query.Where("coalesce(parsed_cv_text, '') <> ''")
	}

	var applications []models.Application
	if err := query.Find(&applications).Error; err != nil {
		serverError(c, "Failed to fetch candidates")
		return
	}

	terms := queryTerms(req.Query)
	results := make([]candidateSearchResult, 0, len(applications))
	for _, app := range applications {
		if result, ok := rankCandidate(app, &req, terms); ok {
			result.Application.ParsedCVText = nil
			results = append(results, result)
		}
	}

	sort.SliceStable(results, func(i, j int) bool { return results[i].MatchScore > results[j].MatchScore })
	if len(results) > req.Limit {
		results = results[:req.Limit]
	}
	c.JSON(http.StatusOK, gin.H{"candidates": results, "count": len(results), "total": len(applications)})
}

type queryTerm struct {
	word    string
	pattern *regexp.Regexp
}

func queryTerms(q string) []queryTerm {
	var terms []queryTerm
	for _, word := range strings.Fields(strings.ToLower(q)) {
		if len(word) > 2 {
			terms = append(terms, queryTerm{word, regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`)})
		}
	}
	return terms
}

func rankCandidate(app models.Application, req *searchCandidatesRequest, terms []queryTerm) (candidateSearchResult, bool) {
	cvText := ""
	if app.ParsedCVText != nil {
		cvText = *app.ParsedCVText
	}
	lower := strings.ToLower(cvText)
	score := 0
	var reasons []string

	if len(terms) > 0 {
		matched := 0
		for _, t := range terms {
			if t.pattern.MatchString(lower) || strings.Contains(lower, t.word) {
				matched++
			}
		}
		if matched == 0 {
			return candidateSearchResult{}, false
		}
		score += matched * 100 / len(terms)
		reasons = append(reasons, fmt.Sprintf("Matched %d/%d search terms", matched, len(terms)))
	}

	var skills []string
	if len(req.Skills) > 0 {
		skills = services.MatchedRequired(cvText, req.Skills)
		if len(skills) > 0 {
			score += len(skills) * 100 / len(req.Skills)
			reasons = append(reasons, fmt.Sprintf("Found %d/%d required skills: %s", len(skills), len(req.Skills), strings.Join(skills, ", ")))
		}
	}
	if req.MinExperience != nil {
		score += 20
		reasons = append(reasons, fmt.Sprintf("Has %d years of experience (required: %d+)", app.YearsOfExperience, *req.MinExperience))
	}
	if len(req.Languages) > 0 {
		if langs := services.ExtractLanguages(cvText, req.Languages); len(langs) > 0 {
			score += len(langs) * 100 / len(req.Languages) / 5
			reasons = append(reasons, "Found languages: "+strings.Join(langs, ", "))
		}
	}
	if req.CurrentPosition != "" {
		score += 15
		reasons = append(reasons, "Current position matches: "+app.CurrentPosition)
	}

	if score == 0 {
		if req.softCriteria() {
			return candidateSearchResult{}, false
		}
		score = 50
	}
	return candidateSearchResult{Application: app, MatchScore: min(score, 100), MatchedSkills: skills, MatchedReasons: reasons}, true
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func GetCandidateDetails(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	app, ok := findCompanyApplication(c, id)
	if !ok {
		return
	}

	cvText := ""
	if app.ParsedCVText != nil {
		cvText = *app.ParsedCVText
	}
	if cvText == "" && (app.ResumeURL != "" || app.ResumePath != nil) {
		if text, err := services.ExtractApplicationCVText(app); err == nil {
			cvText = text
			_ = services.SaveAnalysis(app.ID, text, nil)
		}
	}

	skills, experience := []string{}, 0
	if cvText != "" {
		skills = services.ExtractSkills(cvText, nil)
		experience = services.ExtractExperience(cvText)
	}
	app.ParsedCVText = nil
	c.JSON(http.StatusOK, gin.H{"candidate": app, "cv_text": cvText, "skills": skills, "experience": experience})
}
