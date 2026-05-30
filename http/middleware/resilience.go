package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
	"github.com/ymhhh/goblocks/resilience"
)

// Resilience returns a Gin middleware that applies rate limiting.
func Resilience(policy *resilience.Policy) gin.HandlerFunc {
	return func(c *gin.Context) {
		if policy == nil {
			c.Next()
			return
		}

		if err := policy.Allow(); err != nil {
			if err == resilience.ErrRateLimited {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "rate limit exceeded",
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Next()
	}
}

// ResilienceWithBreaker returns middleware that also checks circuit breaker state.
func ResilienceWithBreaker(policy *resilience.Policy) gin.HandlerFunc {
	return func(c *gin.Context) {
		if policy == nil {
			c.Next()
			return
		}

		if err := policy.Allow(); err != nil {
			status := http.StatusTooManyRequests
			if err == resilience.ErrRateLimited {
				c.AbortWithStatusJSON(status, gin.H{"error": "rate limit exceeded"})
				return
			}
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		if policy.Breaker != nil && policy.Breaker.State() == gobreaker.StateOpen {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "circuit breaker is open",
			})
			return
		}

		c.Next()
	}
}
