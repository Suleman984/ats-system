package models

import (
	"time"

	"github.com/google/uuid"
)

// NurtureCampaign records one automated email sent to a talent-pool candidate.
type NurtureCampaign struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ApplicationID uuid.UUID  `gorm:"type:uuid" json:"application_id"`
	JobID         *uuid.UUID `gorm:"type:uuid" json:"job_id,omitempty"`
	EmailSentAt   time.Time  `json:"email_sent_at"`
	EmailType     string     `json:"email_type"`
	Subject       string     `json:"subject"`
	Status        string     `gorm:"default:sent" json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
}

// NurturePreference stores a candidate's job-alert preferences.
type NurturePreference struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ApplicationID   uuid.UUID  `gorm:"type:uuid" json:"application_id"`
	Email           string     `json:"email"`
	Preferences     *string    `gorm:"type:jsonb" json:"preferences,omitempty"`
	IsActive        bool       `gorm:"default:true" json:"is_active"`
	LastContactedAt *time.Time `json:"last_contacted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
