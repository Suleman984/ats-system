package config

import (
	"os"
	"strings"
)

func IsProduction() bool {
	mode := strings.ToLower(os.Getenv("APP_MODE"))
	return mode == "production" || mode == "prod"
}

func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func FrontendURL() string {
	return strings.TrimRight(GetEnv("FRONTEND_URL", "http://localhost:3000"), "/")
}

// AllowedOrigins reads a comma-separated ALLOWED_ORIGINS, defaulting to FRONTEND_URL.
func AllowedOrigins() []string {
	return splitList(GetEnv("ALLOWED_ORIGINS", FrontendURL()))
}

// TrustedProxies reads a comma-separated TRUSTED_PROXIES list of IPs or CIDRs.
func TrustedProxies() []string {
	return splitList(os.Getenv("TRUSTED_PROXIES"))
}

func splitList(raw string) []string {
	var items []string
	for _, item := range strings.Split(raw, ",") {
		if item = strings.TrimRight(strings.TrimSpace(item), "/"); item != "" {
			items = append(items, item)
		}
	}
	return items
}
