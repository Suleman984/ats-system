package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"unicode"
)

// Criteria defines the shortlisting criteria
type Criteria struct {
	RequiredSkills      []string `json:"required_skills"`
	MinExperience       int      `json:"min_experience"`
	RequiredLanguages   []string `json:"required_languages"`
	MatchJobDescription bool     `json:"match_job_description"`
	JobDescription      string   `json:"job_description,omitempty"`
	JobRequirements     string   `json:"job_requirements,omitempty"`
}

// MatchResult contains the matching analysis results
type MatchResult struct {
	MatchScore      int      `json:"match_score"`      // 0-100 percentage
	Skills          []string `json:"skills"`          // Skills found in CV
	Experience      int      `json:"experience"`       // Years of experience extracted
	Education       string   `json:"education"`       // Education level (for compatibility)
	Languages       []string `json:"languages"`       // Languages found
	Summary         string   `json:"summary"`         // Brief summary
	MatchReason     string   `json:"match_reason"`    // Why matched/not matched
	MissingSkills   []string `json:"missing_skills"`  // Skills from criteria not found
	Strengths       []string `json:"strengths"`       // Key strengths found
	SkillsMatch     int      `json:"skills_match"`    // Skills match percentage
	ExperienceMatch int      `json:"experience_match"` // Experience match percentage
	LanguageMatch   int      `json:"language_match"`  // Language match percentage
}

// ExtractTextFromURL downloads and extracts text from CV file
func ExtractTextFromURL(cvURL string) (string, error) {
	// Download the file
	resp, err := http.Get(cvURL)
	if err != nil {
		return "", fmt.Errorf("failed to download CV: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download CV: status %d", resp.StatusCode)
	}

	// Read file content
	fileBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read CV file: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	
	// For text files
	if strings.Contains(contentType, "text/plain") || strings.HasSuffix(cvURL, ".txt") {
		return string(fileBytes), nil
	}
	
	// For PDFs - extract text (enhanced approach: multiple extraction methods)
	if strings.Contains(contentType, "pdf") || strings.HasSuffix(cvURL, ".pdf") {
		// Enhanced PDF text extraction - try multiple methods
		text := extractTextFromPDF(fileBytes)
		
		// If first method didn't work well, try fallback
		if len(text) < 100 {
			// Try extracting readable text from entire file
			text = extractReadableText(fileBytes)
		}
		
		if len(text) > 100 {
			return text, nil
		}
		// If extraction fails, return error with helpful message
		return "", fmt.Errorf("could not extract text from PDF. Please ensure PDF contains readable text (not scanned images). For scanned PDFs, consider using OCR tools first.")
	}
	
	// For DOCX - try to extract text
	if strings.Contains(contentType, "wordprocessingml") || strings.HasSuffix(cvURL, ".docx") {
		// Simple DOCX text extraction
		text := extractTextFromDOCX(fileBytes)
		if len(text) > 100 {
			return text, nil
		}
	}
	
	// For DOC - try to extract text
	if strings.Contains(contentType, "msword") || strings.HasSuffix(cvURL, ".doc") {
		// DOC files are binary, harder to parse without library
		// Try to extract readable text
		text := extractReadableText(fileBytes)
		if len(text) > 100 {
			return text, nil
		}
	}
	
	// Fallback: try to extract any readable text
	text := extractReadableText(fileBytes)
	if len(text) > 100 {
		return text, nil
	}
	
	return "", fmt.Errorf("could not extract readable text from CV file. Supported formats: PDF (with readable text), DOC, DOCX, TXT. For scanned images, please convert to text first.")
}

// extractTextFromPDF extracts text from PDF bytes (simple approach)
func extractTextFromPDF(data []byte) string {
	// Simple PDF text extraction: look for text streams
	// This is a basic implementation - for production, use a library like pdfcpu
	text := ""
	
	// Look for text between stream markers
	streamStart := []byte("stream")
	streamEnd := []byte("endstream")
	
	startIdx := 0
	for {
		idx := findBytes(data[startIdx:], streamStart)
		if idx == -1 {
			break
		}
		idx += startIdx
		
		endIdx := findBytes(data[idx:], streamEnd)
		if endIdx == -1 {
			break
		}
		endIdx += idx
		
		// Extract text from stream
		streamData := data[idx+len(streamStart) : endIdx]
		text += extractReadableText(streamData) + " "
		
		startIdx = endIdx + len(streamEnd)
	}
	
	return strings.TrimSpace(text)
}

// extractTextFromDOCX extracts text from DOCX (ZIP-based format)
func extractTextFromDOCX(data []byte) string {
	// DOCX is a ZIP file containing XML
	// Look for document.xml content
	// Enhanced approach: extract readable text from XML with better patterns
	text := ""
	
	// Look for text in XML tags - multiple patterns for better extraction
	xmlPatterns := []*regexp.Regexp{
		regexp.MustCompile(`<w:t[^>]*>([^<]+)</w:t>`),           // Standard Word text
		regexp.MustCompile(`<t[^>]*>([^<]+)</t>`),              // Alternative text tag
		regexp.MustCompile(`<text[^>]*>([^<]+)</text>`),        // Generic text tag
		regexp.MustCompile(`xml:space="preserve">([^<]+)</w:t>`), // Preserved space text
	}
	
	for _, pattern := range xmlPatterns {
		matches := pattern.FindAllStringSubmatch(string(data), -1)
		for _, match := range matches {
			if len(match) > 1 {
				text += match[1] + " "
			}
		}
	}
	
	// Also try to extract from raw XML if patterns didn't work well
	if len(text) < 100 {
		// Fallback: extract any readable text between tags
		fallbackPattern := regexp.MustCompile(`>([A-Za-z0-9\s\.\,\:\;\!\?\-]+)<`)
		matches := fallbackPattern.FindAllStringSubmatch(string(data), -1)
		for _, match := range matches {
			if len(match) > 1 && len(match[1]) > 2 {
				text += match[1] + " "
			}
		}
	}
	
	return strings.TrimSpace(text)
}

// extractReadableText extracts readable ASCII/Unicode text from binary data
func extractReadableText(data []byte) string {
	var text strings.Builder
	var word strings.Builder
	
	for _, b := range data {
		// Check if byte is printable ASCII or common Unicode
		if unicode.IsPrint(rune(b)) || b == ' ' || b == '\n' || b == '\r' || b == '\t' {
			if unicode.IsLetter(rune(b)) || unicode.IsDigit(rune(b)) || b == ' ' || b == '-' || b == '_' {
				word.WriteByte(b)
			} else if word.Len() > 0 {
				// End of word
				wordStr := word.String()
				if len(wordStr) > 2 { // Only add words longer than 2 chars
					text.WriteString(wordStr)
					text.WriteByte(' ')
				}
				word.Reset()
			}
		} else if word.Len() > 0 {
			// End of word on non-printable
			wordStr := word.String()
			if len(wordStr) > 2 {
				text.WriteString(wordStr)
				text.WriteByte(' ')
			}
			word.Reset()
		}
	}
	
	// Add last word
	if word.Len() > 2 {
		text.WriteString(word.String())
	}
	
	return strings.TrimSpace(text.String())
}

// findBytes finds byte sequence in data
func findBytes(data []byte, pattern []byte) int {
	if len(pattern) == 0 || len(data) < len(pattern) {
		return -1
	}
	
	for i := 0; i <= len(data)-len(pattern); i++ {
		match := true
		for j := 0; j < len(pattern); j++ {
			if data[i+j] != pattern[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	
	return -1
}

// ExtractExperience extracts years of experience from CV text
func ExtractExperience(cvText string) int {
	// Look for patterns like "5 years", "3+ years", "experience: 4 years"
	experiencePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(\d+)\s*\+?\s*years?\s*(?:of\s*)?experience`),
		regexp.MustCompile(`experience[:\s]+(\d+)\s*years?`),
		regexp.MustCompile(`(\d+)\s*years?\s*experience`),
		regexp.MustCompile(`(\d+)\s*y\.?o\.?e\.?`), // years of experience
	}
	
	maxYears := 0
	cvLower := strings.ToLower(cvText)
	
	for _, pattern := range experiencePatterns {
		matches := pattern.FindAllStringSubmatch(cvLower, -1)
		for _, match := range matches {
			if len(match) > 1 {
				var years int
				fmt.Sscanf(match[1], "%d", &years)
				if years > maxYears {
					maxYears = years
				}
			}
		}
	}
	
	// If no explicit experience found, try to estimate from dates
	if maxYears == 0 {
		// Look for date ranges (e.g., "2020 - 2024", "Jan 2020 - Present")
		dateRangePattern := regexp.MustCompile(`(?:19|20)\d{2}`)
		dates := dateRangePattern.FindAllString(cvText, -1)
		if len(dates) >= 2 {
			var startYear, endYear int
			fmt.Sscanf(dates[0], "%d", &startYear)
			fmt.Sscanf(dates[len(dates)-1], "%d", &endYear)
			if endYear > startYear {
				maxYears = endYear - startYear
			}
		}
	}
	
	return maxYears
}

// ExtractSkills extracts skills from CV text with synonym recognition and inference
func ExtractSkills(cvText string, requiredSkills []string) []string {
	foundSkills := []string{}
	cvLower := strings.ToLower(cvText)
	
	// Extract job title from CV for skills inference
	jobTitle := extractJobTitleFromCV(cvText)
	inferredSkills := InferSkillsFromJobTitle(jobTitle)
	
	// Check each required skill with synonym matching
	for _, skill := range requiredSkills {
		normalizedSkill := NormalizeSkill(skill)
		
		// Direct match
		if strings.Contains(cvLower, normalizedSkill) {
			foundSkills = append(foundSkills, skill)
			continue
		}
		
		// Try word boundary matching
		skillRegex := regexp.MustCompile(`\b` + regexp.QuoteMeta(normalizedSkill) + `\b`)
		if skillRegex.MatchString(cvLower) {
			foundSkills = append(foundSkills, skill)
			continue
		}
		
		// Try synonym matching
		synonyms := GetSkillSynonyms(skill)
		matched := false
		for _, synonym := range synonyms {
			synonymNormalized := NormalizeSkill(synonym)
			if strings.Contains(cvLower, synonymNormalized) {
				foundSkills = append(foundSkills, skill)
				matched = true
				break
			}
			// Try word boundary for synonyms too
			synonymRegex := regexp.MustCompile(`\b` + regexp.QuoteMeta(synonymNormalized) + `\b`)
			if synonymRegex.MatchString(cvLower) {
				foundSkills = append(foundSkills, skill)
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		
		// Try partial match for compound skills (e.g., "React.js" matches "React")
		skillWords := strings.Fields(normalizedSkill)
		if len(skillWords) > 0 {
			firstWord := skillWords[0]
			if len(firstWord) > 3 && strings.Contains(cvLower, firstWord) {
				foundSkills = append(foundSkills, skill)
				continue
			}
		}
		
		// Try skills inference - if required skill matches inferred skills
		for _, inferred := range inferredSkills {
			if strings.Contains(strings.ToLower(inferred), normalizedSkill) || 
			   strings.Contains(normalizedSkill, strings.ToLower(inferred)) {
				foundSkills = append(foundSkills, skill)
				break
			}
		}
	}
	
	// Also extract common skills mentioned in CV (even if not required)
	commonSkills := []string{
		"javascript", "python", "java", "react", "node.js", "sql", "html", "css",
		"typescript", "angular", "vue", "php", "ruby", "go", "rust", "c++", "c#",
		"aws", "docker", "kubernetes", "git", "mongodb", "postgresql", "mysql",
		"agile", "scrum", "api", "rest", "graphql", "microservices",
	}
	
	for _, skill := range commonSkills {
		skillLower := strings.ToLower(skill)
		if strings.Contains(cvLower, skillLower) && !contains(foundSkills, skill) {
			// Check synonyms too
			synonyms := GetSkillSynonyms(skill)
			hasSynonym := false
			for _, synonym := range synonyms {
				if strings.Contains(cvLower, strings.ToLower(synonym)) {
					hasSynonym = true
					break
				}
			}
			if hasSynonym || strings.Contains(cvLower, skillLower) {
				foundSkills = append(foundSkills, skill)
			}
		}
	}
	
	// Add inferred skills from job title (if not already found)
	for _, inferred := range inferredSkills {
		if !contains(foundSkills, inferred) {
			inferredLower := strings.ToLower(inferred)
			if strings.Contains(cvLower, inferredLower) {
				foundSkills = append(foundSkills, inferred)
			}
		}
	}
	
	return foundSkills
}

// extractJobTitleFromCV extracts job title from CV text
func extractJobTitleFromCV(cvText string) string {
	cvLower := strings.ToLower(cvText)
	
	// Look for common job title patterns
	titlePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:position|role|title|job)[:\s]+([a-z\s]+(?:developer|engineer|manager|analyst|designer|lead|architect))`),
		regexp.MustCompile(`(?i)(senior|junior|mid-level)?\s*([a-z\s]+(?:developer|engineer|manager|analyst|designer|lead|architect))`),
		regexp.MustCompile(`(?i)([a-z\s]+(?:developer|engineer|manager|analyst|designer|lead|architect))`),
	}
	
	for _, pattern := range titlePatterns {
		matches := pattern.FindStringSubmatch(cvLower)
		if len(matches) > 1 {
			title := strings.TrimSpace(matches[1])
			if len(title) > 3 && len(title) < 50 {
				return title
			}
		}
	}
	
	return ""
}

// ExtractLanguages extracts languages from CV text
func ExtractLanguages(cvText string, requiredLanguages []string) []string {
	foundLanguages := []string{}
	cvLower := strings.ToLower(cvText)
	
	// Common language patterns
	languagePatterns := map[string][]string{
		"english":    {"english", "fluent english", "native english"},
		"spanish":    {"spanish", "español", "castellano"},
		"french":     {"french", "français"},
		"german":     {"german", "deutsch"},
		"chinese":    {"chinese", "mandarin", "中文"},
		"arabic":     {"arabic", "عربي"},
		"hindi":      {"hindi", "हिंदी"},
		"portuguese": {"portuguese", "português"},
		"italian":    {"italian", "italiano"},
		"japanese":   {"japanese", "日本語"},
	}
	
	// Check required languages
	for _, lang := range requiredLanguages {
		langLower := strings.ToLower(lang)
		if patterns, ok := languagePatterns[langLower]; ok {
			for _, pattern := range patterns {
				if strings.Contains(cvLower, pattern) {
					foundLanguages = append(foundLanguages, lang)
					break
				}
			}
		} else if strings.Contains(cvLower, langLower) {
			foundLanguages = append(foundLanguages, lang)
		}
	}
	
	return foundLanguages
}

// MatchCV analyzes CV against criteria and returns match score
func MatchCV(cvText string, criteria Criteria, jobTitle string) *MatchResult {
	result := &MatchResult{
		Skills:    []string{},
		Languages: []string{},
		Strengths: []string{},
	}
	
	cvLower := strings.ToLower(cvText)
	
	// 1. Extract and match skills (40% weight)
	if len(criteria.RequiredSkills) > 0 {
		result.Skills = ExtractSkills(cvText, criteria.RequiredSkills)
		result.SkillsMatch = (len(result.Skills) * 100) / len(criteria.RequiredSkills)
		if result.SkillsMatch > 100 {
			result.SkillsMatch = 100
		}
		
		// Find missing skills
		for _, skill := range criteria.RequiredSkills {
			if !contains(result.Skills, skill) {
				result.MissingSkills = append(result.MissingSkills, skill)
			}
		}
	} else {
		result.SkillsMatch = 100 // No skills required = full match
	}
	
	// 2. Extract and match experience (30% weight)
	result.Experience = ExtractExperience(cvText)
	if criteria.MinExperience > 0 {
		if result.Experience >= criteria.MinExperience {
			result.ExperienceMatch = 100
		} else if result.Experience > 0 {
			result.ExperienceMatch = (result.Experience * 100) / criteria.MinExperience
		} else {
			result.ExperienceMatch = 0
		}
	} else {
		result.ExperienceMatch = 100 // No experience requirement = full match
	}
	
	// 3. Extract and match languages (20% weight)
	if len(criteria.RequiredLanguages) > 0 {
		result.Languages = ExtractLanguages(cvText, criteria.RequiredLanguages)
		result.LanguageMatch = (len(result.Languages) * 100) / len(criteria.RequiredLanguages)
		if result.LanguageMatch > 100 {
			result.LanguageMatch = 100
		}
	} else {
		result.LanguageMatch = 100 // No language requirement = full match
	}
	
	// 4. Match job description (10% weight)
	jobDescMatch := 0
	if criteria.MatchJobDescription && criteria.JobDescription != "" {
		jobDescLower := strings.ToLower(criteria.JobDescription)
		jobWords := strings.Fields(jobDescLower)
		matchedWords := 0
		
		for _, word := range jobWords {
			if len(word) > 4 && strings.Contains(cvLower, word) {
				matchedWords++
			}
		}
		
		if len(jobWords) > 0 {
			jobDescMatch = (matchedWords * 100) / len(jobWords)
			if jobDescMatch > 100 {
				jobDescMatch = 100
			}
		}
	} else {
		jobDescMatch = 100 // Not required = full match
	}
	
	// Calculate overall match score
	result.MatchScore = (result.SkillsMatch*40 + result.ExperienceMatch*30 + result.LanguageMatch*20 + jobDescMatch*10) / 100
	
	// Generate summary and match reason
	result.Summary = generateSummary(result, criteria)
	result.MatchReason = generateMatchReason(result, criteria)
	
	// Identify strengths
	result.Strengths = identifyStrengths(result, cvText)
	
	return result
}

// MatchCVFromURL analyzes CV from URL against criteria
func MatchCVFromURL(cvURL string, criteria Criteria, jobTitle string) (*MatchResult, error) {
	// Extract text from CV
	cvText, err := ExtractTextFromURL(cvURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract CV text: %w", err)
	}
	
	if len(cvText) < 50 {
		return nil, fmt.Errorf("CV text too short or unreadable")
	}
	
	log.Printf("Extracted %d characters from CV", len(cvText))
	
	// Match against criteria
	result := MatchCV(cvText, criteria, jobTitle)
	
	return result, nil
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func generateSummary(result *MatchResult, criteria Criteria) string {
	parts := []string{}
	
	if result.Experience > 0 {
		parts = append(parts, fmt.Sprintf("%d years experience", result.Experience))
	}
	
	if len(result.Skills) > 0 {
		parts = append(parts, fmt.Sprintf("%d/%d required skills", len(result.Skills), len(criteria.RequiredSkills)))
	}
	
	if len(result.Languages) > 0 {
		parts = append(parts, fmt.Sprintf("%d languages", len(result.Languages)))
	}
	
	if len(parts) > 0 {
		return strings.Join(parts, ", ")
	}
	
	return "Candidate profile analyzed"
}

func generateMatchReason(result *MatchResult, criteria Criteria) string {
	reasons := []string{}
	
	if result.SkillsMatch >= 80 {
		reasons = append(reasons, "Strong skills match")
	} else if result.SkillsMatch >= 50 {
		reasons = append(reasons, "Partial skills match")
	} else if len(criteria.RequiredSkills) > 0 {
		reasons = append(reasons, "Missing key skills")
	}
	
	if result.ExperienceMatch >= 100 {
		reasons = append(reasons, "Meets experience requirement")
	} else if result.ExperienceMatch > 0 {
		reasons = append(reasons, "Below experience requirement")
	}
	
	if result.LanguageMatch >= 100 {
		reasons = append(reasons, "Meets language requirements")
	} else if len(criteria.RequiredLanguages) > 0 {
		reasons = append(reasons, "Missing language requirements")
	}
	
	if len(reasons) > 0 {
		return strings.Join(reasons, ". ")
	}
	
	return "Basic profile match"
}

func identifyStrengths(result *MatchResult, cvText string) []string {
	strengths := []string{}
	
	if result.Experience >= 5 {
		strengths = append(strengths, "Extensive experience")
	}
	
	if len(result.Skills) >= 5 {
		strengths = append(strengths, "Diverse skill set")
	}
	
	if result.SkillsMatch >= 80 {
		strengths = append(strengths, "Strong technical match")
	}
	
	if strings.Contains(strings.ToLower(cvText), "degree") || strings.Contains(strings.ToLower(cvText), "bachelor") || strings.Contains(strings.ToLower(cvText), "master") {
		strengths = append(strengths, "Educational background")
	}
	
	if strings.Contains(strings.ToLower(cvText), "certification") || strings.Contains(strings.ToLower(cvText), "certified") {
		strengths = append(strengths, "Professional certifications")
	}
	
	return strengths
}

