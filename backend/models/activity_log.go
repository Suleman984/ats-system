package models

import (
	"time"

	"github.com/google/uuid"
)

type ActivityLog struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID   *uuid.UUID `gorm:"type:uuid" json:"company_id,omitempty"` // NULL for super admin actions
	AdminID     *uuid.UUID `gorm:"type:uuid" json:"admin_id,omitempty"`   // NULL for system actions
	ActionType  string     `gorm:"size:50;not null" json:"action_type"`  // company_registered, job_created, job_updated, job_deleted, application_shortlisted, etc.
	EntityType  string     `gorm:"size:50;not null" json:"entity_type"`  // company, job, application, etc.
	EntityID    *uuid.UUID `gorm:"type:uuid" json:"entity_id,omitempty"` // ID of the job, application, etc.
	Description string     `gorm:"type:text" json:"description"`        // Human-readable description
	Metadata    *string    `gorm:"type:jsonb" json:"metadata,omitempty"` // Additional details (old values, new values, etc.)
	CreatedAt   time.Time  `json:"created_at"`

	// Relations
	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Admin   *Admin   `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
}

