package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Criteria defines what a job requires from a CV.
type Criteria struct {
	RequiredSkills      []string `json:"required_skills"`
	MinExperience       int      `json:"min_experience"`
	RequiredLanguages   []string `json:"required_languages"`
	MatchJobDescription bool     `json:"match_job_description"`
	JobDescription      string   `json:"job_description,omitempty"`
	JobRequirements     string   `json:"job_requirements,omitempty"`
}

type MatchResult struct {
	MatchScore      int      `json:"match_score"`
	Skills          []string `json:"skills"`
	Experience      int      `json:"experience"`
	Education       string   `json:"education"`
	Languages       []string `json:"languages"`
	Summary         string   `json:"summary"`
	MatchReason     string   `json:"match_reason"`
	MissingSkills   []string `json:"missing_skills"`
	Strengths       []string `json:"strengths"`
	SkillsMatch     int      `json:"skills_match"`
	ExperienceMatch int      `json:"experience_match"`
	LanguageMatch   int      `json:"language_match"`
}

const minCVTextLength = 50

var (
	pdfParenText    = regexp.MustCompile(`\(([^)]+)\)`)
	pdfObjectRef    = regexp.MustCompile(`^\d+\s+\d+\s*R$`)
	pdfObjectHeader = regexp.MustCompile(`^\d+\s*obj$|^endobj$`)
	whitespaceRun   = regexp.MustCompile(`\s+`)
	pdfArtifacts    = []*regexp.Regexp{
		regexp.MustCompile(`\b(?:Pg|P)\s+\d+\s+\d+\s*R\b`),
		regexp.MustCompile(`\b\d+\s+\d+\s*R\b`),
		regexp.MustCompile(`\b\d+(?:\s+\d+)?\s*obj\b`),
		regexp.MustCompile(`\b(?:Type\s+)?StructElem\b`),
		regexp.MustCompile(`\b(?:endobj|endstream|stream)\b`),
		regexp.MustCompile(`/\w+\s+\d+`),
	}
	pdfStructureMarkers = []string{
		"pg ", "endobj", "obj", "type", "structelem", "r ", "stream", "xref", "trailer",
		"startxref", "/filter", "/length", "/w", "/font", "/subtype", "/resources", "/mediabox", "/contents",
	}
	experiencePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(\d+)\s*\+?\s*years?\s*(?:of\s*)?experience`),
		regexp.MustCompile(`experience[:\s]+(\d+)\s*years?`),
		regexp.MustCompile(`(\d+)\s*y\.?o\.?e\.?`),
	}
	yearPattern  = regexp.MustCompile(`(?:19|20)\d{2}`)
	titlePattern = []*regexp.Regexp{
		regexp.MustCompile(`(?:position|role|title|job)[:\s]+([a-z\s]+(?:developer|engineer|manager|analyst|designer|lead|architect))`),
		regexp.MustCompile(`([a-z\s]+(?:developer|engineer|manager|analyst|designer|lead|architect))`),
	}
	commonSkills = []string{
		"javascript", "python", "java", "react", "node.js", "sql", "html", "css",
		"typescript", "angular", "vue", "php", "ruby", "go", "rust", "c++", "c#",
		"aws", "docker", "kubernetes", "git", "mongodb", "postgresql", "mysql",
		"agile", "scrum", "api", "rest", "graphql", "microservices",
	}
	languageAliases = map[string][]string{
		"english":    {"english"},
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
)

// ExtractTextFromURL downloads a CV and extracts its plain text.
func ExtractTextFromURL(cvURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data, contentType, err := download(ctx, cvURL)
	if err != nil {
		return "", err
	}
	text, err := extractText(data, contentType, strings.ToLower(cvURL))
	if err != nil {
		return "", err
	}
	if len(text) < minCVTextLength || !utf8.ValidString(text) {
		return "", errors.New("CV text is too short or unreadable; scanned PDFs need OCR first")
	}
	return text, nil
}

func extractText(data []byte, contentType, name string) (string, error) {
	switch {
	case strings.Contains(contentType, "text/plain") || strings.HasSuffix(name, ".txt"):
		return string(data), nil
	case bytes.HasPrefix(data, []byte("%PDF")):
		return extractPDFText(data), nil
	case bytes.HasPrefix(data, []byte("PK\x03\x04")):
		return extractDOCXText(data)
	default:
		return extractReadableText(data), nil
	}
}

// extractPDFText handles uncompressed text objects and streams; it is heuristic, not a full PDF parser.
func extractPDFText(data []byte) string {
	for _, extract := range []func([]byte) string{extractPDFTextObjects, extractPDFStreams, extractReadableText} {
		if text := cleanPDFText(extract(data)); len(text) > 100 && looksLikeText(text) {
			return text
		}
	}
	return ""
}

func extractPDFTextObjects(data []byte) string {
	var out strings.Builder
	for rest := data; ; {
		start := bytes.Index(rest, []byte("BT"))
		if start == -1 {
			break
		}
		end := bytes.Index(rest[start:], []byte("ET"))
		if end == -1 {
			break
		}
		for _, m := range pdfParenText.FindAllSubmatch(rest[start+2:start+end], -1) {
			if s := string(m[1]); !isPDFStructureMarker(s) && utf8.ValidString(s) && looksLikeText(s) {
				out.WriteString(s)
				out.WriteByte(' ')
			}
		}
		rest = rest[start+end+2:]
	}
	return out.String()
}

func extractPDFStreams(data []byte) string {
	var out strings.Builder
	for offset := 0; ; {
		start := bytes.Index(data[offset:], []byte("stream"))
		if start == -1 {
			break
		}
		start += offset
		end := bytes.Index(data[start:], []byte("endstream"))
		if end == -1 {
			break
		}
		end += start

		header := string(data[max(0, start-200):start])
		compressed := strings.Contains(header, "/FlateDecode") || strings.Contains(header, "/LZWDecode") || strings.Contains(header, "/DCTDecode")
		if !compressed {
			if text := cleanPDFText(extractReadableText(data[start+len("stream") : end])); len(text) > 10 {
				out.WriteString(text)
				out.WriteByte(' ')
			}
		}
		offset = end + len("endstream")
	}
	return out.String()
}

func isPDFStructureMarker(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	for _, marker := range pdfStructureMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return pdfObjectRef.MatchString(s) || pdfObjectHeader.MatchString(s)
}

func cleanPDFText(text string) string {
	for _, pattern := range pdfArtifacts {
		text = pattern.ReplaceAllString(text, " ")
	}
	return strings.TrimSpace(whitespaceRun.ReplaceAllString(text, " "))
}

// looksLikeText rejects binary noise: at least 30% letters and either a common word or enough length.
func looksLikeText(text string) bool {
	letters, visible := 0, 0
	for _, r := range text {
		if unicode.IsLetter(r) {
			letters++
		}
		if !unicode.IsSpace(r) {
			visible++
		}
	}
	if visible == 0 || letters*100/visible < 30 {
		return false
	}
	lower := strings.ToLower(text)
	for _, word := range []string{"the", "and", "or", "is", "are", "was", "be", "have", "has"} {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return len(text) > 200 && letters > 50
}

// extractDOCXText reads word/document.xml from the DOCX zip and joins its text runs.
func extractDOCXText(data []byte) (string, error) {
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("invalid DOCX: %w", err)
	}
	idx := slices.IndexFunc(archive.File, func(f *zip.File) bool { return f.Name == "word/document.xml" })
	if idx == -1 {
		return "", errors.New("invalid DOCX: missing word/document.xml")
	}
	rc, err := archive.File[idx].Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	var out strings.Builder
	decoder := xml.NewDecoder(io.LimitReader(rc, maxDownloadBytes))
	inText := false
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("invalid DOCX XML: %w", err)
		}
		switch t := token.(type) {
		case xml.StartElement:
			inText = t.Name.Local == "t"
		case xml.EndElement:
			if t.Name.Local == "p" {
				out.WriteByte('\n')
			}
			inText = false
		case xml.CharData:
			if inText {
				out.Write(t)
			}
		}
	}
	return strings.TrimSpace(out.String()), nil
}

// extractReadableText keeps printable ASCII words longer than two characters.
func extractReadableText(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	var text, word strings.Builder
	valid := 0
	flush := func() {
		if word.Len() > 2 {
			text.WriteString(word.String())
			text.WriteByte(' ')
		}
		word.Reset()
	}
	for _, b := range data {
		r := rune(b)
		if b < 0x80 && (unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune(" -_.,", r)) {
			word.WriteByte(b)
			valid++
		} else {
			flush()
		}
	}
	flush()

	if valid*100/len(data) < 5 {
		return ""
	}
	return strings.TrimSpace(text.String())
}

func ExtractExperience(cvText string) int {
	lower := strings.ToLower(cvText)
	years := 0
	for _, pattern := range experiencePatterns {
		for _, m := range pattern.FindAllStringSubmatch(lower, -1) {
			if n, err := strconv.Atoi(m[1]); err == nil && n < 60 && n > years {
				years = n
			}
		}
	}
	if years > 0 {
		return years
	}

	// Fall back to the span between the first and last year mentioned.
	dates := yearPattern.FindAllString(cvText, -1)
	if len(dates) >= 2 {
		first, _ := strconv.Atoi(dates[0])
		last, _ := strconv.Atoi(dates[len(dates)-1])
		if last > first {
			return last - first
		}
	}
	return 0
}

// containsWord matches whole words only, so "go" does not match inside "django".
func containsWord(text, word string) bool {
	if word == "" {
		return false
	}
	for offset := 0; offset < len(text); {
		idx := strings.Index(text[offset:], word)
		if idx == -1 {
			return false
		}
		start := offset + idx
		end := start + len(word)
		if isWordBoundary(text, start-1) && isWordBoundary(text, end) {
			return true
		}
		offset = start + 1
	}
	return false
}

func isWordBoundary(text string, i int) bool {
	if i < 0 || i >= len(text) {
		return true
	}
	b := text[i]
	return !(b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || b >= 'A' && b <= 'Z')
}

// MatchedRequired returns the required skills present in a CV.
func MatchedRequired(cvText string, required []string) []string {
	found := ExtractSkills(cvText, required)
	var matched []string
	for _, skill := range required {
		if containsFold(found, skill) {
			matched = append(matched, skill)
		}
	}
	return matched
}

// ExtractSkills matches required skills (with synonyms and title inference) plus common skills.
func ExtractSkills(cvText string, requiredSkills []string) []string {
	lower := strings.ToLower(cvText)
	inferred := InferSkillsFromJobTitle(extractJobTitle(lower))
	var found []string

	for _, skill := range requiredSkills {
		normalized := NormalizeSkill(skill)
		matched := containsWord(lower, normalized) || containsWord(lower, strings.ToLower(skill))
		for _, synonym := range GetSkillSynonyms(skill) {
			if matched {
				break
			}
			matched = containsWord(lower, NormalizeSkill(synonym)) || containsWord(lower, strings.ToLower(synonym))
		}
		if !matched {
			if words := strings.Fields(normalized); len(words) > 0 && len(words[0]) > 3 {
				matched = containsWord(lower, words[0])
			}
		}
		for _, skillFromTitle := range inferred {
			if matched {
				break
			}
			s := strings.ToLower(skillFromTitle)
			matched = containsWord(s, normalized) || containsWord(normalized, s)
		}
		if matched {
			found = append(found, skill)
		}
	}

	for _, skill := range append(slices.Clone(commonSkills), inferred...) {
		if containsWord(lower, strings.ToLower(skill)) && !containsFold(found, skill) {
			found = append(found, skill)
		}
	}
	return found
}

func extractJobTitle(lowerCV string) string {
	for _, pattern := range titlePattern {
		if m := pattern.FindStringSubmatch(lowerCV); len(m) > 1 {
			if title := strings.TrimSpace(m[1]); len(title) > 3 && len(title) < 50 {
				return title
			}
		}
	}
	return ""
}

func ExtractLanguages(cvText string, requiredLanguages []string) []string {
	lower := strings.ToLower(cvText)
	var found []string
	for _, lang := range requiredLanguages {
		key := strings.ToLower(lang)
		aliases, ok := languageAliases[key]
		if !ok {
			aliases = []string{key}
		}
		if slices.ContainsFunc(aliases, func(a string) bool { return strings.Contains(lower, a) }) {
			found = append(found, lang)
		}
	}
	return found
}

// MatchCV scores a CV: skills 40%, experience 30%, languages 20%, job description 10%.
func MatchCV(cvText string, criteria Criteria) *MatchResult {
	lower := strings.ToLower(cvText)
	result := &MatchResult{Skills: []string{}, Languages: []string{}, MissingSkills: []string{}}

	result.SkillsMatch = 100
	if n := len(criteria.RequiredSkills); n > 0 {
		result.Skills = ExtractSkills(cvText, criteria.RequiredSkills)
		matched := 0
		for _, skill := range criteria.RequiredSkills {
			if containsFold(result.Skills, skill) {
				matched++
			} else {
				result.MissingSkills = append(result.MissingSkills, skill)
			}
		}
		result.SkillsMatch = matched * 100 / n
	}

	result.Experience = ExtractExperience(cvText)
	result.ExperienceMatch = 100
	if criteria.MinExperience > 0 {
		result.ExperienceMatch = min(100, result.Experience*100/criteria.MinExperience)
	}

	result.LanguageMatch = 100
	if n := len(criteria.RequiredLanguages); n > 0 {
		result.Languages = ExtractLanguages(cvText, criteria.RequiredLanguages)
		result.LanguageMatch = len(result.Languages) * 100 / n
	}

	descMatch := 100
	if criteria.MatchJobDescription && criteria.JobDescription != "" {
		words := strings.Fields(strings.ToLower(criteria.JobDescription))
		hits := 0
		for _, w := range words {
			if len(w) > 4 && strings.Contains(lower, w) {
				hits++
			}
		}
		descMatch = hits * 100 / len(words)
	}

	result.MatchScore = (result.SkillsMatch*40 + result.ExperienceMatch*30 + result.LanguageMatch*20 + descMatch*10) / 100
	result.Summary = summarize(result, criteria)
	result.MatchReason = matchReason(result, criteria)
	result.Strengths = strengths(result, lower)
	return result
}

func containsFold(list []string, item string) bool {
	return slices.ContainsFunc(list, func(s string) bool { return strings.EqualFold(s, item) })
}

func summarize(r *MatchResult, c Criteria) string {
	var parts []string
	if r.Experience > 0 {
		parts = append(parts, fmt.Sprintf("%d years experience", r.Experience))
	}
	if len(c.RequiredSkills) > 0 {
		parts = append(parts, fmt.Sprintf("%d/%d required skills", len(c.RequiredSkills)-len(r.MissingSkills), len(c.RequiredSkills)))
	}
	if len(r.Languages) > 0 {
		parts = append(parts, fmt.Sprintf("%d languages", len(r.Languages)))
	}
	if len(parts) == 0 {
		return "Candidate profile analyzed"
	}
	return strings.Join(parts, ", ")
}

func matchReason(r *MatchResult, c Criteria) string {
	var reasons []string
	switch {
	case r.SkillsMatch >= 80:
		reasons = append(reasons, "Strong skills match")
	case r.SkillsMatch >= 50:
		reasons = append(reasons, "Partial skills match")
	case len(c.RequiredSkills) > 0:
		reasons = append(reasons, "Missing key skills")
	}
	switch {
	case r.ExperienceMatch >= 100:
		reasons = append(reasons, "Meets experience requirement")
	case r.ExperienceMatch > 0:
		reasons = append(reasons, "Below experience requirement")
	}
	if r.LanguageMatch >= 100 {
		reasons = append(reasons, "Meets language requirements")
	} else if len(c.RequiredLanguages) > 0 {
		reasons = append(reasons, "Missing language requirements")
	}
	if len(reasons) == 0 {
		return "Basic profile match"
	}
	return strings.Join(reasons, ". ")
}

func strengths(r *MatchResult, lowerCV string) []string {
	out := []string{}
	if r.Experience >= 5 {
		out = append(out, "Extensive experience")
	}
	if len(r.Skills) >= 5 {
		out = append(out, "Diverse skill set")
	}
	if r.SkillsMatch >= 80 {
		out = append(out, "Strong technical match")
	}
	if strings.Contains(lowerCV, "degree") || strings.Contains(lowerCV, "bachelor") || strings.Contains(lowerCV, "master") {
		out = append(out, "Educational background")
	}
	if strings.Contains(lowerCV, "certification") || strings.Contains(lowerCV, "certified") {
		out = append(out, "Professional certifications")
	}
	return out
}
