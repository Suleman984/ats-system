package models

import (
	"time"

	"github.com/google/uuid"
)

// Billing: plans define limits, subscriptions record paid periods, payments record Safepay checkouts.

type SubscriptionPlan struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name            string    `json:"name"`
	Price           float64   `json:"price"`
	Currency        string    `gorm:"default:USD" json:"currency"`
	Features        string    `gorm:"type:jsonb" json:"features"`
	MaxJobs         int       `json:"max_jobs"`
	MaxApplications int       `json:"max_applications"`
	MaxAdmins       int       `json:"max_admins"`
	AIShortlisting  bool      `json:"ai_shortlisting"`
	PricePKR        int       `gorm:"column:price_pkr" json:"price_pkr"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Subscription struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CompanyID            uuid.UUID `gorm:"type:uuid" json:"company_id"`
	PlanID               uuid.UUID `gorm:"type:uuid" json:"plan_id"`
	Status               string    `gorm:"default:active" json:"status"`
	CurrentPeriodStart   time.Time `json:"current_period_start"`
	CurrentPeriodEnd     time.Time `json:"current_period_end"`
	CancelAtPeriodEnd    bool      `json:"cancel_at_period_end"`
	StripeSubscriptionID *string   `json:"stripe_subscription_id,omitempty"`
	PayPalSubscriptionID *string   `gorm:"column:paypal_subscription_id" json:"paypal_subscription_id,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	Plan *SubscriptionPlan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

type Payment struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CompanyID        uuid.UUID  `gorm:"type:uuid" json:"company_id"`
	SubscriptionID   *uuid.UUID `gorm:"type:uuid" json:"subscription_id,omitempty"`
	Amount           float64    `json:"amount"`
	Currency         string     `gorm:"default:USD" json:"currency"`
	Status           string     `gorm:"default:pending" json:"status"`
	PaymentMethod    string     `json:"payment_method"`
	PaymentGatewayID *string    `json:"payment_gateway_id,omitempty"`
	TransactionID    string     `json:"transaction_id"`
	Metadata         *string    `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
