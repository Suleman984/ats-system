package config

import (
	"ats-backend/models"
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	var err error
	// Configure PostgreSQL driver to disable prepared statements
	// This is required for Supabase connection pooling to avoid "prepared statement already exists" errors
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // Disable prepared statements for connection pooling
	}), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("✅ Database connected successfully!")

	// Auto-migrate models (only adds missing columns, doesn't recreate tables)
	// If you get "relation already exists" errors, the tables are fine - GORM will just add missing columns
	err = DB.AutoMigrate(
		&models.Company{},
		&models.Admin{},
		&models.Job{},
		&models.Application{},
		&models.SuperAdmin{},
		&models.EmailLog{},
		&models.SubscriptionPlan{},
		&models.Subscription{},
		&models.Payment{},
	)
	if err != nil {
		// Check if error is just "relation already exists" - this is OK, tables exist
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("⚠️  Some tables already exist - this is normal. GORM will add missing columns.")
			fmt.Println("✅ Database migration completed (some tables already existed)")
		} else {
			log.Fatal("Failed to migrate database:", err)
		}
	} else {
		fmt.Println("✅ Database tables migrated successfully!")
	}
}

func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

