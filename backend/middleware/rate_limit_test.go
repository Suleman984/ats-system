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

func TestAuthMiddlewareRequiresSessionCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", AuthMiddleware(), func(c *gin.Context) { c.Status(http.StatusOK) })

	for _, cookie := range []string{"", "not-a-jwt"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if cookie != "" {
			req.AddCookie(&http.Cookie{Name: AdminCookie, Value: cookie})
		}
		// A bearer header alone no longer authenticates.
		req.Header.Set("Authorization", "Bearer abc")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("cookie %q: status %d, want 401", cookie, rec.Code)
		}
	}
}

func TestAuthMiddlewareRequiresClientHeaderOnWrites(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/", AuthMiddleware(), func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: AdminCookie, Value: "anything"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST without %s: status %d, want 403", ClientHeader, rec.Code)
	}
}
