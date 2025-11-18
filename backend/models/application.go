package models

import (
	"time"

	"github.com/google/uuid"
)

type Application struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	JobID              uuid.UUID  `gorm:"type:uuid;not null" json:"job_id"`
	FullName           string     `gorm:"size:255;not null" json:"full_name"`
	Email              string     `gorm:"size:255;not null" json:"email"`
	Phone              string     `gorm:"size:50" json:"phone"`
	ResumeURL          string     `gorm:"type:text;not null" json:"resume_url"`
	CoverLetter        string     `gorm:"type:text" json:"cover_letter"`
	YearsOfExperience  int        `json:"years_of_experience"`
	CurrentPosition    string     `gorm:"size:255" json:"current_position"`
	LinkedinURL        string     `gorm:"size:255" json:"linkedin_url"`
	PortfolioURL       string     `gorm:"size:255" json:"portfolio_url"`
	Status             string     `gorm:"size:50;default:'pending'" json:"status"`
	Score              int        `gorm:"default:0" json:"score"` // AI match score 0-100
	AnalysisResult     *string    `gorm:"type:jsonb" json:"analysis_result,omitempty"` // AI analysis JSON
	ParsedCVText       *string    `gorm:"type:text" json:"parsed_cv_text,omitempty"` // Extracted CV text for searching
	AppliedAt          time.Time  `json:"applied_at"`
	ReviewedAt         *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy         *uuid.UUID `gorm:"type:uuid" json:"reviewed_by,omitempty"`

	// Relations
	Job Job `gorm:"foreignKey:JobID" json:"job,omitempty"`
}

