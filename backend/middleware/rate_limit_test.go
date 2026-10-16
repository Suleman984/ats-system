package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimitBlocksAfterLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", RateLimit(2, time.Minute), func(c *gin.Context) { c.Status(http.StatusOK) })

	codes := make([]int, 3)
	for i := range codes {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		codes[i] = rec.Code
	}
	if codes[0] != http.StatusOK || codes[1] != http.StatusOK || codes[2] != http.StatusTooManyRequests {
		t.Fatalf("status codes = %v, want [200 200 429]", codes)
	}
}

func TestAuthMiddlewareRequiresBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", AuthMiddleware(), func(c *gin.Context) { c.Status(http.StatusOK) })

	for _, header := range []string{"", "Token abc", "Bearer "} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("header %q: status %d, want 401", header, rec.Code)
		}
	}
}
