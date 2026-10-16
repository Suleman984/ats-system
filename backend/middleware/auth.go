package middleware

import (
	"net/http"
	"strings"

	"ats-backend/utils"

	"github.com/gin-gonic/gin"
)

const (
	CtxAdminID      = "admin_id"
	CtxCompanyID    = "company_id"
	CtxSuperAdminID = "super_admin_id"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c)
		if !ok {
			return
		}
		adminID, companyID, err := utils.VerifyJWT(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		c.Set(CtxAdminID, adminID)
		c.Set(CtxCompanyID, companyID)
		c.Next()
	}
}

func SuperAdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c)
		if !ok {
			return
		}
		superAdminID, err := utils.VerifySuperAdminJWT(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		c.Set(CtxSuperAdminID, superAdminID)
		c.Next()
	}
}

func bearerToken(c *gin.Context) (string, bool) {
	token, found := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
	if !found || token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return "", false
	}
	return token, true
}
