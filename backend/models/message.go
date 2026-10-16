package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	SenderCandidate = "candidate"
	SenderRecruiter = "recruiter"
)

type Message struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ApplicationID uuid.UUID  `gorm:"type:uuid" json:"application_id"`
	SenderType    string     `json:"sender_type"`
	SenderID      *uuid.UUID `gorm:"type:uuid" json:"sender_id,omitempty"`
	SenderEmail   string     `json:"sender_email"`
	Message       string     `json:"message"`
	IsRead        bool       `json:"is_read"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
