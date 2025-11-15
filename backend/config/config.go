package config

import (
	"os"
	"strings"
)

// AppMode represents the application mode
type AppMode string

const (
	ModeDevelopment AppMode = "development"
	ModeProduction  AppMode = "production"
)

var (
	// CurrentMode is the current application mode
	CurrentMode AppMode
)

// InitConfig initializes application configuration
func InitConfig() {
	mode := strings.ToLower(os.Getenv("APP_MODE"))
	switch mode {
	case "production", "prod":
		CurrentMode = ModeProduction
	default:
		CurrentMode = ModeDevelopment
	}
}

// IsDevelopment returns true if in development mode
func IsDevelopment() bool {
	return CurrentMode == ModeDevelopment
}

// IsProduction returns true if in production mode
func IsProduction() bool {
	return CurrentMode == ModeProduction
}


// GetRequiredEnv gets required environment variable or panics
func GetRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Required environment variable not set: " + key)
	}
	return value
}

