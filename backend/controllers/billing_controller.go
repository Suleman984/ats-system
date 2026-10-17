package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetBilling returns the plan catalogue, the company's current plan and usage, and recent payments.
func GetBilling(c *gin.Context) {
	company := companyID(c)
	var plans []models.SubscriptionPlan
	if err := config.DB.Order("price_pkr").Find(&plans).Error; err != nil {
		serverError(c, "Failed to load plans")
		return
	}
	current, err := services.CompanyPlan(company)
	if err != nil {
		serverError(c, "Failed to load current plan")
		return
	}
	usage, err := services.CompanyUsage(company)
	if err != nil {
		serverError(c, "Failed to load usage")
		return
	}
	var subscription models.Subscription
	config.DB.Where("company_id = ?", company).Order("current_period_end desc").Limit(1).Find(&subscription)
	var payments []models.Payment
	config.DB.Where("company_id = ?", company).Order("created_at desc").Limit(10).Find(&payments)

	body := gin.H{
		"plans": plans, "current_plan": current, "usage": usage, "payments": payments,
		"checkout_available": services.SafepayConfigured(),
	}
	if subscription.ID != uuid.Nil {
		body["subscription"] = subscription
	}
	c.JSON(http.StatusOK, body)
}

type paymentMetadata struct {
	PlanID uuid.UUID `json:"plan_id"`
}

// StartCheckout creates a pending payment for one month of a plan and returns the Safepay checkout URL.
func StartCheckout(c *gin.Context) {
	var req struct {
		Plan string `json:"plan" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	var plan models.SubscriptionPlan
	if err := config.DB.Where("lower(name) = ?", strings.ToLower(req.Plan)).First(&plan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}
	if plan.PricePKR <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This plan has no price configured"})
		return
	}
	if !services.SafepayConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Online payments aren't set up yet"})
		return
	}

	meta, _ := json.Marshal(paymentMetadata{PlanID: plan.ID})
	metaStr := string(meta)
	payment := models.Payment{
		CompanyID: companyID(c), Amount: float64(plan.PricePKR), Currency: "PKR", Status: "pending",
		PaymentMethod: "safepay", TransactionID: uuid.NewString(), Metadata: &metaStr,
	}
	if err := config.DB.Create(&payment).Error; err != nil {
		serverError(c, "Failed to start checkout")
		return
	}

	returnURL := config.FrontendURL() + "/admin/dashboard/billing?payment=" + payment.ID.String()
	tracker, checkoutURL, err := services.CreateSafepayCheckout(payment.TransactionID, int64(plan.PricePKR)*100, returnURL, returnURL+"&cancelled=1")
	if err != nil {
		log.Printf("safepay checkout for payment %s: %v", payment.ID, err)
		config.DB.Model(&payment).Update("status", "failed")
		c.JSON(http.StatusBadGateway, gin.H{"error": "Couldn't reach the payment provider. Please try again."})
		return
	}
	config.DB.Model(&payment).Update("payment_gateway_id", tracker)
	c.JSON(http.StatusCreated, gin.H{"checkout_url": checkoutURL, "payment_id": payment.ID})
}

// ConfirmPayment is called when the buyer returns from checkout, in case the webhook hasn't arrived yet.
func ConfirmPayment(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	var payment models.Payment
	if err := config.DB.Where("id = ? AND company_id = ?", id, companyID(c)).First(&payment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}
	if err := settlePayment(&payment); err != nil {
		log.Printf("confirm payment %s: %v", payment.ID, err)
	}
	c.JSON(http.StatusOK, gin.H{"payment": payment})
}

// SafepayWebhook handles payment events. The body only tells us which tracker to re-check;
// settlePayment asks Safepay for the real status before activating anything.
func SafepayWebhook(c *gin.Context) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil || !services.VerifySafepayWebhook(body, c.GetHeader("X-SFPY-SIGNATURE")) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}
	var event struct {
		Type string `json:"type"`
		Data struct {
			Tracker string `json:"tracker"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &event); err != nil || event.Data.Tracker == "" {
		c.JSON(http.StatusOK, gin.H{"ignored": true})
		return
	}
	var payment models.Payment
	if err := config.DB.Where("payment_gateway_id = ?", event.Data.Tracker).First(&payment).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"ignored": true})
		return
	}
	if err := settlePayment(&payment); err != nil {
		log.Printf("webhook %s for payment %s: %v", event.Type, payment.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "retry later"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"received": true})
}

// settlePayment checks a pending payment with Safepay and, if paid in full, extends the subscription.
// It is idempotent: the row lock and status check stop a webhook and a confirm from both activating.
func settlePayment(payment *models.Payment) error {
	if payment.Status != "pending" || payment.PaymentGatewayID == nil {
		return nil
	}
	state, amount, currency, err := services.SafepayTrackerStatus(*payment.PaymentGatewayID)
	if err != nil {
		return err
	}
	if state != services.SafepayTrackerPaid {
		return nil
	}
	if amount != int64(payment.Amount)*100 || !strings.EqualFold(currency, payment.Currency) {
		log.Printf("payment %s: Safepay reports %d %s, expected %d %s; not activating", payment.ID, amount, currency, int64(payment.Amount)*100, payment.Currency)
		return config.DB.Model(payment).Update("status", "failed").Error
	}

	var meta paymentMetadata
	if payment.Metadata == nil || json.Unmarshal([]byte(*payment.Metadata), &meta) != nil {
		return errors.New("payment has no plan metadata")
	}
	return config.DB.Transaction(func(tx *gorm.DB) error {
		var locked models.Payment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, "id = ?", payment.ID).Error; err != nil {
			return err
		}
		if locked.Status != "pending" {
			*payment = locked
			return nil
		}
		var plan models.SubscriptionPlan
		if err := tx.First(&plan, "id = ?", meta.PlanID).Error; err != nil {
			return err
		}

		// Renewing early stacks the new month after the current period instead of overlapping it.
		start := time.Now()
		var latest models.Subscription
		tx.Where("company_id = ? AND plan_id = ?", payment.CompanyID, plan.ID).Order("current_period_end desc").Limit(1).Find(&latest)
		if latest.ID != uuid.Nil && latest.CurrentPeriodEnd.After(start) {
			start = latest.CurrentPeriodEnd
		}
		sub := models.Subscription{
			CompanyID: payment.CompanyID, PlanID: plan.ID, Status: "active",
			CurrentPeriodStart: start, CurrentPeriodEnd: start.AddDate(0, 1, 0),
		}
		if err := tx.Create(&sub).Error; err != nil {
			return err
		}
		if err := tx.Model(&locked).Updates(map[string]any{"status": "completed", "subscription_id": sub.ID}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Company{}).Where("id = ?", payment.CompanyID).
			Updates(map[string]any{"subscription_tier": strings.ToLower(plan.Name), "subscription_status": models.CompanyActive}).Error; err != nil {
			return err
		}
		services.LogActivity(&payment.CompanyID, nil, "subscription_paid", "company", &payment.CompanyID,
			"Paid for "+plan.Name+" until "+sub.CurrentPeriodEnd.Format("2 Jan 2006"),
			map[string]any{"plan": plan.Name, "amount_pkr": payment.Amount, "payment_id": payment.ID})
		locked.Status, locked.SubscriptionID = "completed", &sub.ID
		*payment = locked
		return nil
	})
}
