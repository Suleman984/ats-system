package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending     = "pending"
	StatusCVViewed    = "cv_viewed"
	StatusShortlisted = "shortlisted"
	StatusRejected    = "rejected"
)

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

	Job *Job `gorm:"foreignKey:JobID" json:"job,omitempty"`
}

// JobTitle falls back to a placeholder once the job has been deleted.
func (a *Application) JobTitle() string {
	if a.Job != nil && a.Job.ID != uuid.Nil {
		return a.Job.Title
	}
	return "Unknown Job (Job Deleted)"
}
