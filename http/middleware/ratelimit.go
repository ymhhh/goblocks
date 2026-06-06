package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
	"github.com/ymhhh/goblocks/metrics"
	"github.com/ymhhh/goblocks/resilience"
)

const httpProtocol = "http"

// GlobalRateLimit applies L1 service-wide rate limiting.
func GlobalRateLimit(policy *resilience.Policy, m *metrics.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		if policy == nil {
			c.Next()
			return
		}
		if err := policy.AllowGlobal(c.Request.Context()); err != nil {
			if err == resilience.ErrRateLimited {
				if m != nil {
					m.RecordRateLimitRejected(httpProtocol, string(resilience.ScopeGlobal))
				}
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
				return
			}
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.Next()
	}
}

// RateLimitByKey applies rate limiting for a custom key (L2 user or L3 route).
func RateLimitByKey(
	keyFn func(*gin.Context) string,
	rule resilience.LimitRule,
	scope resilience.Scope,
	policy *resilience.Policy,
	m *metrics.Registry,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		if policy == nil || policy.RateLimits.Backend == nil {
			c.Next()
			return
		}
		key := keyFn(c)
		if key == "" {
			c.Next()
			return
		}
		ok, err := policy.RateLimits.Backend.Allow(c.Request.Context(), key, rule)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		if !ok {
			if m != nil {
				m.RecordRateLimitRejected(httpProtocol, string(scope))
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

// UserRateLimit applies L2 per-user rate limiting using context user id.
func UserRateLimit(policy *resilience.Policy, m *metrics.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		if policy == nil || !policy.RateLimits.UserEnabled {
			c.Next()
			return
		}
		if err := policy.AllowUser(c.Request.Context(), ""); err != nil {
			if err == resilience.ErrRateLimited {
				if m != nil {
					m.RecordRateLimitRejected(httpProtocol, string(resilience.ScopeUser))
				}
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
				return
			}
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.Next()
	}
}

// RouteRateLimit applies L3 rate limiting when a route rule exists in config.
func RouteRateLimit(policy *resilience.Policy, m *metrics.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		if policy == nil {
			c.Next()
			return
		}
		p := routePath(c)
		if p == "" {
			c.Next()
			return
		}
		if err := policy.AllowRoute(c.Request.Context(), c.Request.Method, p); err != nil {
			if err == resilience.ErrRateLimited {
				if m != nil {
					m.RecordRateLimitRejected(httpProtocol, string(resilience.ScopeRoute))
				}
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
				return
			}
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.Next()
	}
}

// BreakerCheck rejects requests when the circuit breaker is open.
func BreakerCheck(policy *resilience.Policy, m *metrics.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		if policy == nil {
			c.Next()
			return
		}
		if policy.Breaker != nil && policy.Breaker.State() == gobreaker.StateOpen {
			if m != nil {
				m.RecordCircuitBreakerRejected(httpProtocol)
			}
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "circuit breaker is open"})
			return
		}
		c.Next()
	}
}

// Resilience returns a Gin middleware that applies rate limiting.
// Deprecated: use GlobalRateLimit and BreakerCheck instead.
func Resilience(policy *resilience.Policy, m *metrics.Registry) gin.HandlerFunc {
	return GlobalRateLimit(policy, m)
}

// ResilienceWithBreaker returns middleware with global rate limit and breaker check.
// Deprecated: use GlobalRateLimit, UserRateLimit, RouteRateLimit, and BreakerCheck.
func ResilienceWithBreaker(policy *resilience.Policy, m *metrics.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		if policy == nil {
			c.Next()
			return
		}
		if err := policy.AllowGlobal(c.Request.Context()); err != nil {
			if err == resilience.ErrRateLimited {
				if m != nil {
					m.RecordRateLimitRejected(httpProtocol, string(resilience.ScopeGlobal))
				}
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
				return
			}
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		if policy.Breaker != nil && policy.Breaker.State() == gobreaker.StateOpen {
			if m != nil {
				m.RecordCircuitBreakerRejected(httpProtocol)
			}
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "circuit breaker is open"})
			return
		}
		c.Next()
	}
}

// GinUserKey returns user rate-limit key from gin context (set via resilience.ContextWithUserID on request context).
func GinUserKey(c *gin.Context) string {
	return resilience.UserKeyFromContext(c.Request.Context())
}

// GinContextWithUserID stores user id on the request context.
func GinContextWithUserID(c *gin.Context, userID string) {
	c.Request = c.Request.WithContext(resilience.ContextWithUserID(c.Request.Context(), userID))
}

func routePath(c *gin.Context) string {
	return c.FullPath()
}
