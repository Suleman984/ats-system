package models

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionPlan defines available pricing plans
type SubscriptionPlan struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"` // e.g., "Free", "Basic", "Pro", "Enterprise"
	Price       float64   `gorm:"not null" json:"price"`          // Monthly price
	Currency    string    `gorm:"size:10;default:'USD'" json:"currency"`
	Features    string    `gorm:"type:jsonb" json:"features"` // JSON array of features
	MaxJobs     int       `gorm:"default:5" json:"max_jobs"`   // Maximum jobs per month
	MaxApplications int   `gorm:"default:100" json:"max_applications"` // Max applications per month
	AIShortlisting bool   `gorm:"default:false" json:"ai_shortlisting"` // AI features enabled
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Subscription tracks company subscriptions
type Subscription struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID       uuid.UUID  `gorm:"type:uuid;not null" json:"company_id"`
	PlanID          uuid.UUID  `gorm:"type:uuid;not null" json:"plan_id"`
	Status          string     `gorm:"size:50;default:'active'" json:"status"` // active, cancelled, expired, trial
	CurrentPeriodStart time.Time `json:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
	CancelAtPeriodEnd  bool      `gorm:"default:false" json:"cancel_at_period_end"`
	StripeSubscriptionID *string `gorm:"size:255" json:"stripe_subscription_id,omitempty"`
	PayPalSubscriptionID *string `gorm:"size:255" json:"paypal_subscription_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	
	// Relations
	Company Company         `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Plan    SubscriptionPlan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

// Payment tracks payment transactions
type Payment struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID       uuid.UUID  `gorm:"type:uuid;not null" json:"company_id"`
	SubscriptionID  *uuid.UUID `gorm:"type:uuid" json:"subscription_id,omitempty"`
	Amount          float64    `gorm:"not null" json:"amount"`
	Currency        string     `gorm:"size:10;default:'USD'" json:"currency"`
	Status          string     `gorm:"size:50;default:'pending'" json:"status"` // pending, completed, failed, refunded
	PaymentMethod   string     `gorm:"size:50" json:"payment_method"` // stripe, paypal, easypaisa, jazzcash, bank_transfer
	PaymentGatewayID *string   `gorm:"size:255" json:"payment_gateway_id,omitempty"` // External payment ID
	TransactionID   string     `gorm:"size:255" json:"transaction_id"` // Our internal transaction ID
	Metadata        *string    `gorm:"type:jsonb" json:"metadata,omitempty"` // Additional payment data
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	
	// Relations
	Company Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}

