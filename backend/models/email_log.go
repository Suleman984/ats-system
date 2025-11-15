package models

import (
	"time"

	"github.com/google/uuid"
)

type EmailLog struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ApplicationID *uuid.UUID `gorm:"type:uuid" json:"application_id,omitempty"`
	EmailType     string     `gorm:"size:50" json:"email_type"`
	SentTo        string     `gorm:"size:255" json:"sent_to"`
	SentAt        time.Time  `json:"sent_at"`
	Status        string     `gorm:"size:50" json:"status"`
}

