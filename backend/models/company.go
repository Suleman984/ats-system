package models

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyName       string    `gorm:"size:255;not null" json:"company_name"`
	Email             string    `gorm:"size:255;unique;not null" json:"email"`
	CompanyWebsite    string    `gorm:"size:255" json:"company_website"`
	EmbeddedMode      bool      `gorm:"default:false" json:"embedded_mode"` // Use embedded dashboard
	EmbedDomain       *string   `gorm:"size:255" json:"embed_domain,omitempty"` // Allowed domain for embedding
	SubscriptionStatus string   `gorm:"size:50;default:'trial'" json:"subscription_status"`
	SubscriptionTier  string   `gorm:"size:50;default:'starter'" json:"subscription_tier"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

