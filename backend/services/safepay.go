package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Safepay hosted checkout (https://apidocs.getsafepay.com). Env: SAFEPAY_ENV (sandbox|production),
// SAFEPAY_PUBLIC_KEY, SAFEPAY_SECRET_KEY, SAFEPAY_WEBHOOK_SECRET.

var safepayClient = &http.Client{Timeout: 20 * time.Second}

// SafepayTrackerPaid is the tracker state Safepay reports once a payment has been captured.
const SafepayTrackerPaid = "TRACKER_ENDED"

func safepayProduction() bool { return os.Getenv("SAFEPAY_ENV") == "production" }

func safepayAPIBase() string {
	if safepayProduction() {
		return "https://api.getsafepay.com"
	}
	return "https://sandbox.api.getsafepay.com"
}

func safepayCheckoutBase() string {
	if safepayProduction() {
		return "https://getsafepay.com"
	}
	return "https://sandbox.api.getsafepay.com"
}

func SafepayConfigured() bool {
	return os.Getenv("SAFEPAY_PUBLIC_KEY") != "" && os.Getenv("SAFEPAY_SECRET_KEY") != ""
}

func safepayRequest(method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequest(method, safepayAPIBase()+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-SFPY-MERCHANT-SECRET", os.Getenv("SAFEPAY_SECRET_KEY"))

	resp, err := safepayClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("safepay %s %s returned %d: %s", method, path, resp.StatusCode, truncate(raw, 300))
	}
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}

func truncate(b []byte, n int) string {
	if len(b) > n {
		return string(b[:n]) + "…"
	}
	return string(b)
}

// CreateSafepayCheckout starts a one-time PKR payment and returns its tracker and hosted checkout URL.
// amountPaisa is in the currency's smallest unit (rupees × 100).
func CreateSafepayCheckout(orderID string, amountPaisa int64, redirectURL, cancelURL string) (tracker, checkoutURL string, err error) {
	if !SafepayConfigured() {
		return "", "", errors.New("safepay keys are not configured")
	}
	params := map[string]any{
		"merchant_api_key": os.Getenv("SAFEPAY_PUBLIC_KEY"),
		"intent":           "CYBERSOURCE",
		"mode":             "payment",
		"currency":         "PKR",
		"amount":           amountPaisa,
	}
	var created struct {
		Data struct {
			Tracker struct {
				Token string `json:"token"`
			} `json:"tracker"`
		} `json:"data"`
	}
	// The trailing slash is required; Safepay answers 405 without it.
	if err := safepayRequest(http.MethodPost, "/order/payments/v3/", params, &created); err != nil {
		return "", "", err
	}
	tracker = created.Data.Tracker.Token
	if tracker == "" {
		return "", "", errors.New("safepay returned no tracker token")
	}

	// Metadata comes back in webhooks; failure here shouldn't block checkout.
	_ = safepayRequest(http.MethodPost, "/order/payments/v3/"+url.PathEscape(tracker)+"/metadata",
		map[string]any{"data": map[string]string{"source": "ats", "order_id": orderID}}, nil)

	var tbt struct {
		Data string `json:"data"`
	}
	if err := safepayRequest(http.MethodPost, "/client/passport/v1/token", params, &tbt); err != nil {
		return "", "", err
	}

	env := "sandbox"
	if safepayProduction() {
		env = "production"
	}
	q := url.Values{
		"environment":  {env},
		"tracker":      {tracker},
		"tbt":          {tbt.Data},
		"source":       {"custom"},
		"order_id":     {orderID},
		"redirect_url": {redirectURL},
		"cancel_url":   {cancelURL},
	}
	return tracker, safepayCheckoutBase() + "/embedded/?" + q.Encode(), nil
}

// SafepayTrackerStatus asks Safepay directly whether a tracker was paid; this, not the browser redirect
// or webhook body, is what unlocks a plan.
func SafepayTrackerStatus(tracker string) (state string, amountPaisa int64, currency string, err error) {
	var out struct {
		Data struct {
			Tracker struct {
				State          string `json:"state"`
				PurchaseTotals struct {
					BaseAmount struct {
						Amount   int64  `json:"amount"`
						Currency string `json:"currency"`
					} `json:"base_amount"`
				} `json:"purchase_totals"`
			} `json:"tracker"`
		} `json:"data"`
	}
	err = safepayRequest(http.MethodGet, "/order/payments/v3/"+url.PathEscape(tracker)+"/status", nil, &out)
	t := out.Data.Tracker
	return t.State, t.PurchaseTotals.BaseAmount.Amount, t.PurchaseTotals.BaseAmount.Currency, err
}

// VerifySafepayWebhook checks the X-SFPY-SIGNATURE header: lowercase-hex HMAC-SHA512 of the body with the
// webhook secret. Safepay's own SDKs sign re-encoded JSON, so the compacted body is accepted too.
func VerifySafepayWebhook(body []byte, signature string) bool {
	secret := os.Getenv("SAFEPAY_WEBHOOK_SECRET")
	if secret == "" || signature == "" {
		return false
	}
	want, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	candidates := [][]byte{body}
	var compact bytes.Buffer
	if json.Compact(&compact, body) == nil {
		candidates = append(candidates, compact.Bytes())
	}
	for _, payload := range candidates {
		mac := hmac.New(sha512.New, []byte(secret))
		mac.Write(payload)
		if hmac.Equal(mac.Sum(nil), want) {
			return true
		}
	}
	return false
}
