package controllers

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestSafeFileName(t *testing.T) {
	cases := map[string]string{
		"cv.pdf":                          "cv.pdf",
		"../../etc/passwd":                "passwd",
		`C:\Users\me\My CV.docx`:          "My CV.docx",
		"bad\x00\nname\".pdf":             "badname.pdf",
		strings.Repeat("é", 300) + ".pdf": strings.Repeat("é", 196) + ".pdf",
	}
	for in, want := range cases {
		if got := safeFileName(in); got != want {
			t.Errorf("safeFileName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCandidateFieldsRequireOneCVSource(t *testing.T) {
	id := uuid.New()
	cases := []struct {
		name string
		f    candidateFields
		ok   bool
	}{
		{"link only", candidateFields{ResumeURL: "https://example.com/cv.pdf"}, true},
		{"upload only", candidateFields{ResumeUploadID: &id}, true},
		{"neither", candidateFields{}, false},
		{"both", candidateFields{ResumeURL: "https://example.com/cv.pdf", ResumeUploadID: &id}, false},
		{"javascript link", candidateFields{ResumeURL: "javascript:alert(1)"}, false},
		{"portfolio link and upload", candidateFields{ResumeUploadID: &id, PortfolioURL: "https://x.dev", PortfolioUploadID: &id}, false},
	}
	for _, tc := range cases {
		if err := tc.f.validateLinks(); (err == nil) != tc.ok {
			t.Errorf("%s: err = %v, want ok=%v", tc.name, err, tc.ok)
		}
	}
}
