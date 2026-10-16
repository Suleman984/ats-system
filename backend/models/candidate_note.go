package models

import (
	"time"

	"github.com/google/uuid"
)

type CandidateNote struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ApplicationID uuid.UUID `gorm:"type:uuid" json:"application_id"`
	AdminID       uuid.UUID `gorm:"type:uuid" json:"admin_id"`
	Note          string    `json:"note"`
	IsPrivate     bool      `json:"is_private"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Admin *Admin `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
}
