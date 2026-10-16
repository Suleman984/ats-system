package models

import (
	"time"

	"github.com/google/uuid"
)

type ActivityLog struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CompanyID   *uuid.UUID `gorm:"type:uuid" json:"company_id,omitempty"`
	AdminID     *uuid.UUID `gorm:"type:uuid" json:"admin_id,omitempty"`
	ActionType  string     `json:"action_type"`
	EntityType  string     `json:"entity_type"`
	EntityID    *uuid.UUID `gorm:"type:uuid" json:"entity_id,omitempty"`
	Description string     `json:"description"`
	Metadata    *string    `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`

	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Admin   *Admin   `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
}
