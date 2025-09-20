package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"urlshorty/internal/rate"
)

// RateLimit enforces a simple per-IP token bucket for the current route.
func RateLimit(lim *rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !lim.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limited"})
			return
		}
		c.Next()
	}
}
