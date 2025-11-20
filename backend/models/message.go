package models

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a message in the candidate communication system
type Message struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ApplicationID uuid.UUID  `gorm:"type:uuid;not null" json:"application_id"`
	SenderType    string     `gorm:"size:20;not null" json:"sender_type"` // "candidate" or "recruiter"
	SenderID      *uuid.UUID `gorm:"type:uuid" json:"sender_id,omitempty"` // Admin ID if recruiter, null if candidate
	SenderEmail   string     `gorm:"size:255;not null" json:"sender_email"` // Email of sender
	Message       string     `gorm:"type:text;not null" json:"message"`
	IsRead        bool       `gorm:"default:false" json:"is_read"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`

	// Relations
	Application Application `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`
}

