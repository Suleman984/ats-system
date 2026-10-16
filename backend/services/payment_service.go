package services

import (
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
)

// Payment scaffolding for the planned subscription plans. Only bank transfer is functional;
// the card and wallet gateways return errNotImplemented until their SDKs are integrated.

type PaymentProvider string

const (
	ProviderStripe       PaymentProvider = "stripe"
	ProviderPayPal       PaymentProvider = "paypal"
	ProviderEasyPaisa    PaymentProvider = "easypaisa"
	ProviderJazzCash     PaymentProvider = "jazzcash"
	ProviderBankTransfer PaymentProvider = "bank_transfer"
)

var errNotImplemented = errors.New("payment provider integration not implemented yet")

type PaymentRequest struct {
	Amount      float64         `json:"amount"`
	Currency    string          `json:"currency"`
	CompanyID   string          `json:"company_id"`
	PlanID      string          `json:"plan_id"`
	Provider    PaymentProvider `json:"provider"`
	Description string          `json:"description"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
}

type PaymentResponse struct {
	Success       bool   `json:"success"`
	TransactionID string `json:"transaction_id"`
	PaymentURL    string `json:"payment_url,omitempty"`
	PaymentID     string `json:"payment_id,omitempty"`
	Message       string `json:"message,omitempty"`
}

// GetAvailableProviders lists the providers whose credentials are configured.
func GetAvailableProviders() []PaymentProvider {
	var providers []PaymentProvider
	if os.Getenv("STRIPE_SECRET_KEY") != "" {
		providers = append(providers, ProviderStripe)
	}
	if os.Getenv("PAYPAL_CLIENT_ID") != "" && os.Getenv("PAYPAL_SECRET") != "" {
		providers = append(providers, ProviderPayPal)
	}
	if os.Getenv("EASYPAISA_MERCHANT_ID") != "" && os.Getenv("EASYPAISA_PASSWORD") != "" {
		providers = append(providers, ProviderEasyPaisa)
	}
	if os.Getenv("JAZZCASH_MERCHANT_ID") != "" && os.Getenv("JAZZCASH_PASSWORD") != "" {
		providers = append(providers, ProviderJazzCash)
	}
	return append(providers, ProviderBankTransfer)
}

func ProcessPayment(req PaymentRequest) (*PaymentResponse, error) {
	switch req.Provider {
	case ProviderBankTransfer:
		return &PaymentResponse{
			Success:       true,
			TransactionID: "BANK_" + uuid.NewString(),
			Message:       "Payment pending manual verification. Please transfer funds and upload proof.",
		}, nil
	case ProviderStripe, ProviderPayPal, ProviderEasyPaisa, ProviderJazzCash:
		return &PaymentResponse{Message: string(req.Provider) + " integration pending"}, errNotImplemented
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", req.Provider)
	}
}

func VerifyPayment(provider PaymentProvider, paymentID string) (bool, error) {
	switch provider {
	case ProviderStripe, ProviderPayPal, ProviderEasyPaisa, ProviderJazzCash:
		return false, errNotImplemented
	default:
		return false, fmt.Errorf("verification not supported for provider: %s", provider)
	}
}
