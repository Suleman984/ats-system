package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type window struct {
	start time.Time
	count int
}

// RateLimit allows `limit` requests per client IP in each fixed `period`.
func RateLimit(limit int, period time.Duration) gin.HandlerFunc {
	var (
		mu      sync.Mutex
		clients = map[string]*window{}
	)

	go func() {
		for range time.Tick(period) {
			mu.Lock()
			for ip, w := range clients {
				if time.Since(w.start) > period {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		w, ok := clients[ip]
		if !ok || now.Sub(w.start) > period {
			w = &window{start: now}
			clients[ip] = w
		}
		w.count++
		exceeded := w.count > limit
		mu.Unlock()

		if exceeded {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests, please try again later"})
			return
		}
		c.Next()
	}
}
