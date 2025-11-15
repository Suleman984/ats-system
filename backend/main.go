package main

import (
	"ats-backend/config"
	"ats-backend/routes"
	"ats-backend/services"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database
	config.InitDB()

	// Initialize Supabase Storage buckets (optional - can be created manually)
	if err := services.CreateBucketIfNotExists("resumes", true); err != nil {
		log.Printf("Warning: Failed to create resumes bucket: %v", err)
		log.Printf("Note: You can create buckets manually in Supabase Dashboard → Storage")
	}
	if err := services.CreateBucketIfNotExists("portfolios", true); err != nil {
		log.Printf("Warning: Failed to create portfolios bucket: %v", err)
		log.Printf("Note: You can create buckets manually in Supabase Dashboard → Storage")
	}

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware())

	// Setup routes
	routes.SetupRoutes(router)

	// Start server
	port := config.GetEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	router.Run(":" + port)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

