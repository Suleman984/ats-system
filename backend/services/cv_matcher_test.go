package services

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

const sampleCV = `Jane Doe - Senior Software Engineer
I have 6 years of experience building web platforms with React, Node.js and PostgreSQL.
Fluent in English and Spanish. Bachelor degree in Computer Science. AWS certified.`

func TestExtractExperience(t *testing.T) {
	cases := map[string]int{
		"I have 6 years of experience":   6,
		"Experience: 3 years in support": 3,
		"Worked 2015 - 2021 at Acme":     6,
		"No numbers here at all":         0,
	}
	for text, want := range cases {
		if got := ExtractExperience(text); got != want {
			t.Errorf("ExtractExperience(%q) = %d, want %d", text, got, want)
		}
	}
}

func TestMatchCVScoresAndMissingSkills(t *testing.T) {
	result := MatchCV(sampleCV, Criteria{
		RequiredSkills:    []string{"React", "Node.js", "Kotlin"},
		MinExperience:     5,
		RequiredLanguages: []string{"English"},
	})

	if result.ExperienceMatch != 100 || result.LanguageMatch != 100 {
		t.Fatalf("experience=%d language=%d, want 100/100", result.ExperienceMatch, result.LanguageMatch)
	}
	if len(result.MissingSkills) != 1 || result.MissingSkills[0] != "Kotlin" {
		t.Fatalf("missing skills = %v, want [Kotlin]", result.MissingSkills)
	}
	if result.SkillsMatch != 66 {
		t.Fatalf("skills match = %d, want 66", result.SkillsMatch)
	}
	if result.MatchScore <= 0 || result.MatchScore > 100 {
		t.Fatalf("match score out of range: %d", result.MatchScore)
	}
}

func TestMatchCVWithoutCriteriaIsFullMatch(t *testing.T) {
	if got := MatchCV(sampleCV, Criteria{}).MatchScore; got != 100 {
		t.Fatalf("score = %d, want 100", got)
	}
}

func TestExtractDOCXText(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("word/document.xml")
	w.Write([]byte(`<w:document xmlns:w="w"><w:body>
		<w:p><w:r><w:t>Senior Go Engineer</w:t></w:r></w:p>
		<w:p><w:r><w:t xml:space="preserve">Five years </w:t></w:r><w:r><w:t>with Kubernetes</w:t></w:r></w:p>
	</w:body></w:document>`))
	zw.Close()

	text, err := extractDOCXText(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Senior Go Engineer") || !strings.Contains(text, "Five years with Kubernetes") {
		t.Fatalf("unexpected DOCX text: %q", text)
	}
}

func TestExtractTextRoutesByContent(t *testing.T) {
	text, err := extractText([]byte("plain resume text"), "text/plain", "cv.txt")
	if err != nil || text != "plain resume text" {
		t.Fatalf("got %q, %v", text, err)
	}
}

func TestSkillsMatchWholeWordsOnly(t *testing.T) {
	cv := "Python and Django developer who writes good code"
	if got := MatchedRequired(cv, []string{"Go", "Python"}); len(got) != 1 || got[0] != "Python" {
		t.Fatalf("matched = %v, want [Python]", got)
	}
}
