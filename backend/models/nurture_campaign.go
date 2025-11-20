package models

import (
	"time"

	"github.com/google/uuid"
)

// NurtureCampaign represents an automated email sent to a candidate
type NurtureCampaign struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ApplicationID uuid.UUID  `gorm:"type:uuid;not null" json:"application_id"`
	JobID       *uuid.UUID `gorm:"type:uuid" json:"job_id,omitempty"`
	EmailSentAt time.Time  `gorm:"not null" json:"email_sent_at"`
	EmailType   string     `gorm:"size:50;not null" json:"email_type"` // 'job_alert', 'check_in', 'opportunity'
	Subject     string     `gorm:"size:255" json:"subject"`
	Status      string     `gorm:"size:50;default:'sent'" json:"status"` // 'sent', 'opened', 'clicked', 'bounced'
	CreatedAt   time.Time  `json:"created_at"`

	// Relations
	Application Application `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`
	Job         *Job        `gorm:"foreignKey:JobID" json:"job,omitempty"`
}

// NurturePreference stores candidate preferences for job alerts
type NurturePreference struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ApplicationID   uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_app_email" json:"application_id"`
	Email           string     `gorm:"size:255;not null;uniqueIndex:idx_app_email" json:"email"`
	Preferences     *string    `gorm:"type:jsonb" json:"preferences,omitempty"` // Job preferences, location, salary range, etc.
	IsActive        bool       `gorm:"default:true" json:"is_active"`
	LastContactedAt *time.Time `json:"last_contacted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relations
	Application Application `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`
}

