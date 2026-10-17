package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	StatusPending     = "pending"
	StatusCVViewed    = "cv_viewed"
	StatusShortlisted = "shortlisted"
	StatusInterview   = "interview"
	StatusOffer       = "offer"
	StatusHired       = "hired"
	StatusRejected    = "rejected"
)

// NextStage is the stage a recruiter can advance each post-shortlist status to.
var NextStage = map[string]string{
	StatusShortlisted: StatusInterview,
	StatusInterview:   StatusOffer,
	StatusOffer:       StatusHired,
}

// Undecided reports whether no shortlist/reject decision has been made yet.
func Undecided(status string) bool {
	return status == StatusPending || status == StatusCVViewed
}

type Application struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	JobID                *uuid.UUID `gorm:"type:uuid" json:"job_id,omitempty"`
	CompanyID            uuid.UUID  `gorm:"type:uuid" json:"company_id"`
	FullName             string     `json:"full_name"`
	Email                string     `json:"email"`
	Phone                string     `json:"phone"`
	ResumeURL            string     `json:"resume_url"`
	CoverLetter          string     `json:"cover_letter"`
	YearsOfExperience    int        `json:"years_of_experience"`
	CurrentPosition      string     `json:"current_position"`
	LinkedinURL          string     `json:"linkedin_url"`
	PortfolioURL         string     `json:"portfolio_url"`
	Status               string     `gorm:"default:pending" json:"status"`
	Score                int        `json:"score"`
	AnalysisResult       *string    `gorm:"type:jsonb" json:"analysis_result,omitempty"`
	ParsedCVText         *string    `json:"parsed_cv_text,omitempty"`
	AppliedAt            time.Time  `json:"applied_at"`
	ReviewedAt           *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy           *uuid.UUID `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	CVViewedAt           *time.Time `json:"cv_viewed_at,omitempty"`
	CVViewedBy           *uuid.UUID `gorm:"type:uuid" json:"cv_viewed_by,omitempty"`
	ExpectedResponseDate *time.Time `gorm:"type:date" json:"expected_response_date,omitempty"`
	LastStatusUpdate     *time.Time `json:"last_status_update,omitempty"`

	ReferralSource    string     `json:"referral_source,omitempty"`
	ReferredByName    string     `json:"referred_by_name,omitempty"`
	ReferredByEmail   string     `json:"referred_by_email,omitempty"`
	ReferredByPhone   string     `json:"referred_by_phone,omitempty"`
	InTalentPool      bool       `json:"in_talent_pool"`
	TalentPoolAddedAt *time.Time `json:"talent_pool_added_at,omitempty"`
	TalentPoolAddedBy *uuid.UUID `gorm:"type:uuid" json:"talent_pool_added_by,omitempty"`

	// Object paths in the private buckets; never serialized. Clients see only the Has* flags and
	// fetch files through signed URLs.
	ResumePath       *string `json:"-"`
	PortfolioPath    *string `json:"-"`
	HasResumeFile    bool    `gorm:"-" json:"has_resume_file"`
	HasPortfolioFile bool    `gorm:"-" json:"has_portfolio_file"`
	// Candidate messages the company hasn't read yet; filled in by list endpoints.
	UnreadMessages int `gorm:"-" json:"unread_messages"`

	Job *Job `gorm:"foreignKey:JobID" json:"job,omitempty"`
}

func (a *Application) AfterFind(*gorm.DB) error {
	a.SetFileFlags()
	return nil
}

func (a *Application) SetFileFlags() {
	a.HasResumeFile, a.HasPortfolioFile = a.ResumePath != nil, a.PortfolioPath != nil
}

// Upload kinds and the bucket each is stored in.
const (
	UploadResume    = "resume"
	UploadPortfolio = "portfolio"
)

// Upload is a file a candidate sent before submitting; an application claims it exactly once.
type Upload struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Kind         string
	ObjectPath   string
	ContentType  string
	SizeBytes    int
	OriginalName string
	ParsedText   *string
	ClaimedAt    *time.Time
	CreatedAt    time.Time
}

// JobTitle falls back to a placeholder once the job has been deleted.
func (a *Application) JobTitle() string {
	if a.Job != nil && a.Job.ID != uuid.Nil {
		return a.Job.Title
	}
	return "Unknown Job (Job Deleted)"
}
