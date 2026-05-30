package metrics

import (
	"time"
)

// ObserveAIRequest records AI chat completion metrics.
func (r *Registry) ObserveAIRequest(model, result string, duration time.Duration) {
	if r == nil {
		return
	}
	if model == "" {
		model = "unknown"
	}
	r.AIRequestsTotal.WithLabelValues(model, result).Inc()
	r.AIRequestDuration.WithLabelValues(model).Observe(duration.Seconds())
}
