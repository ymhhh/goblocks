package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
)

// HTTPMiddleware records HTTP request count and latency.
func (r *Registry) HTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if r == nil {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		method := c.Request.Method
		status := c.Writer.Status()

		r.HTTPRequestsTotal.WithLabelValues(method, path, statusLabel(status)).Inc()
		r.HTTPRequestDuration.WithLabelValues(method, path).Observe(time.Since(start).Seconds())
	}
}
