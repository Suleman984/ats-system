package models

import (
	"time"

	"github.com/google/uuid"
)

// CandidateNote represents a note added by a recruiter about a candidate
type CandidateNote struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ApplicationID uuid.UUID  `gorm:"type:uuid;not null" json:"application_id"`
	AdminID       uuid.UUID  `gorm:"type:uuid;not null" json:"admin_id"`
	Note          string     `gorm:"type:text;not null" json:"note"`
	IsPrivate     bool       `gorm:"default:false" json:"is_private"` // Private notes only visible to creator
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// Relations
	Application Application `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`
	Admin       Admin       `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
}

