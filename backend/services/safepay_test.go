package services

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"testing"
)

func sign(t *testing.T, secret string, body []byte) string {
	t.Helper()
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySafepayWebhook(t *testing.T) {
	t.Setenv("SAFEPAY_WEBHOOK_SECRET", "whsec_test")
	raw := []byte("{\n  \"type\": \"payment.succeeded\",\n  \"data\": {\"tracker\": \"track_1\"}\n}")
	compact := []byte(`{"type":"payment.succeeded","data":{"tracker":"track_1"}}`)

	cases := []struct {
		name string
		sig  string
		want bool
	}{
		{"raw body", sign(t, "whsec_test", raw), true},
		{"re-encoded body, as Safepay's SDKs sign it", sign(t, "whsec_test", compact), true},
		{"wrong secret", sign(t, "other", raw), false},
		{"tampered body", sign(t, "whsec_test", []byte(`{"type":"payment.succeeded","data":{"tracker":"track_2"}}`)), false},
		{"not hex", "zz", false},
		{"missing", "", false},
	}
	for _, tc := range cases {
		if got := VerifySafepayWebhook(raw, tc.sig); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestVerifySafepayWebhookWithoutSecret(t *testing.T) {
	t.Setenv("SAFEPAY_WEBHOOK_SECRET", "")
	body := []byte(`{}`)
	if VerifySafepayWebhook(body, sign(t, "", body)) {
		t.Fatal("must reject when no webhook secret is configured")
	}
}
