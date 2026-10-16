package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"ats-backend/config"
	"ats-backend/routes"
	"ats-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	if err := utils.InitJWT(); err != nil {
		log.Fatal(err)
	}
	if err := config.InitDB(); err != nil {
		log.Fatalf("database: %v", err)
	}

	router := gin.Default()
	// Only trust X-Forwarded-For from listed proxies so rate limits can't be spoofed.
	if err := router.SetTrustedProxies(config.TrustedProxies()); err != nil {
		log.Fatalf("TRUSTED_PROXIES: %v", err)
	}
	router.Use(corsMiddleware(config.AllowedOrigins()))
	routes.SetupRoutes(router)

	srv := &http.Server{
		Addr:              ":" + config.GetEnv("PORT", "8080"),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("API listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func corsMiddleware(allowed []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && slices.Contains(allowed, origin) {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			h.Set("Access-Control-Max-Age", "86400")
			h.Add("Vary", "Origin")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
