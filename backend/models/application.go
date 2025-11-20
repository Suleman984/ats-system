package models

import (
	"time"

	"github.com/google/uuid"
)

type Application struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	JobID              *uuid.UUID `gorm:"type:uuid" json:"job_id,omitempty"` // Nullable - job may be deleted
	CompanyID          uuid.UUID  `gorm:"type:uuid;not null" json:"company_id"` // Added to track company even if job deleted
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
	CVViewedAt         *time.Time `json:"cv_viewed_at,omitempty"` // When recruiter first viewed CV
	CVViewedBy         *uuid.UUID `gorm:"type:uuid" json:"cv_viewed_by,omitempty"` // Admin who viewed CV
	ExpectedResponseDate *time.Time `gorm:"type:date" json:"expected_response_date,omitempty"` // Expected response date
	LastStatusUpdate   *time.Time `json:"last_status_update,omitempty"` // Last time status was updated
	
	// CRM Fields
	ReferralSource     string     `gorm:"size:255" json:"referral_source,omitempty"` // How they heard about the job
	ReferredByName     string     `gorm:"size:255" json:"referred_by_name,omitempty"` // Name of referrer
	ReferredByEmail    string     `gorm:"size:255" json:"referred_by_email,omitempty"` // Email of referrer
	ReferredByPhone    string     `gorm:"size:50" json:"referred_by_phone,omitempty"` // Phone of referrer
	InTalentPool       bool       `gorm:"default:false" json:"in_talent_pool"` // Marked for future opportunities
	TalentPoolAddedAt  *time.Time `json:"talent_pool_added_at,omitempty"` // When added to talent pool
	TalentPoolAddedBy  *uuid.UUID `gorm:"type:uuid" json:"talent_pool_added_by,omitempty"` // Admin who added to talent pool

	// Relations
	Job Job `gorm:"foreignKey:JobID" json:"job,omitempty"`
}

