package middleware

import (
	"ats-backend/utils"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
			c.Abort()
			return
		}

		token := parts[1]

		// Verify token
		claims, err := utils.VerifyJWT(token)
		if err != nil {
			log.Printf("AuthMiddleware: Token verification failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Log extracted claims for debugging
		log.Printf("AuthMiddleware: Extracted claims - AdminID: '%s', CompanyID: '%s'", claims.AdminID, claims.CompanyID)

		// Validate claims are not empty
		if claims.AdminID == "" {
			log.Printf("AuthMiddleware: ERROR - AdminID is empty in token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: missing admin ID"})
			c.Abort()
			return
		}

		if claims.CompanyID == "" {
			log.Printf("AuthMiddleware: ERROR - CompanyID is empty in token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: missing company ID"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("admin_id", claims.AdminID)
		c.Set("company_id", claims.CompanyID)

		// Verify what was set
		log.Printf("AuthMiddleware: Set context - admin_id: '%s', company_id: '%s'", claims.AdminID, claims.CompanyID)

		c.Next()
	}
}

