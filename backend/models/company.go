package models

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CompanyName        string    `json:"company_name"`
	Email              string    `json:"email"`
	CompanyWebsite     string    `json:"company_website"`
	EmbeddedMode       bool      `json:"embedded_mode"`
	EmbedDomain        *string   `json:"embed_domain,omitempty"`
	SubscriptionStatus string    `gorm:"default:trial" json:"subscription_status"`
	SubscriptionTier   string    `gorm:"default:starter" json:"subscription_tier"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
