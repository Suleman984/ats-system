package models

import (
	"time"

	"github.com/google/uuid"
)

// Admin roles: owners manage the team, billing and destructive actions; recruiters run hiring.
const (
	RoleOwner     = "owner"
	RoleRecruiter = "recruiter"
)

type Admin struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CompanyID    uuid.UUID `gorm:"type:uuid" json:"company_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `gorm:"default:owner" json:"role"`
	CreatedAt    time.Time `json:"created_at"`

	PasswordChangedAt *time.Time `json:"-"`
	EmailVerifiedAt   *time.Time `json:"-"`
}

// EmailVerificationToken stores only the SHA-256 of the emailed token.
type EmailVerificationToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdminID   uuid.UUID `gorm:"type:uuid"`
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// PasswordResetToken stores only the SHA-256 of the emailed token.
type PasswordResetToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdminID   uuid.UUID `gorm:"type:uuid"`
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type SuperAdmin struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// TeamInvite is a pending invitation; only the SHA-256 of the emailed token is stored.
type TeamInvite struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CompanyID  uuid.UUID  `gorm:"type:uuid" json:"company_id"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	TokenHash  string     `json:"-"`
	InvitedBy  *uuid.UUID `gorm:"type:uuid" json:"invited_by,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
