package models

import (
	"time"

	"github.com/google/uuid"
)

// EmailLog is reserved for delivery tracking; nothing writes to it yet.
type EmailLog struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ApplicationID *uuid.UUID `gorm:"type:uuid" json:"application_id,omitempty"`
	EmailType     string     `json:"email_type"`
	SentTo        string     `json:"sent_to"`
	SentAt        time.Time  `json:"sent_at"`
	Status        string     `json:"status"`
}
