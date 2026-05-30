package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sony/gobreaker"
)

const namespace = "goblocks"

// Registry holds Prometheus collectors for goblocks.
type Registry struct {
	promRegistry *prometheus.Registry

	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec

	GRPCRequestsTotal   *prometheus.CounterVec
	GRPCRequestDuration *prometheus.HistogramVec

	RateLimitRejectedTotal      *prometheus.CounterVec
	CircuitBreakerRejectedTotal *prometheus.CounterVec
	CircuitBreakerState         *prometheus.GaugeVec

	AIRequestsTotal   *prometheus.CounterVec
	AIRequestDuration *prometheus.HistogramVec
}

// NewRegistry creates a Registry with a dedicated prometheus registry.
func NewRegistry() *Registry {
	promRegistry := prometheus.NewRegistry()
	r := &Registry{promRegistry: promRegistry}

	r.HTTPRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests processed.",
	}, []string{"method", "path", "status"})

	r.HTTPRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request latency in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path"})

	r.GRPCRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "grpc_server_handled_total",
		Help:      "Total number of gRPC requests handled.",
	}, []string{"grpc_method", "grpc_code"})

	r.GRPCRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "grpc_server_handling_seconds",
		Help:      "gRPC request handling latency in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"grpc_method"})

	r.RateLimitRejectedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "resilience_rate_limit_rejected_total",
		Help:      "Total number of requests rejected by rate limiting.",
	}, []string{"protocol"})

	r.CircuitBreakerRejectedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "resilience_circuit_breaker_rejected_total",
		Help:      "Total number of requests rejected by circuit breaker.",
	}, []string{"protocol"})

	r.CircuitBreakerState = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "resilience_circuit_breaker_state",
		Help:      "Circuit breaker state: 0=closed, 1=half-open, 2=open.",
	}, []string{"name"})

	r.AIRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "ai_requests_total",
		Help:      "Total number of AI chat completion requests.",
	}, []string{"model", "status"})

	r.AIRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "ai_request_duration_seconds",
		Help:      "AI chat completion latency in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"model"})

	promRegistry.MustRegister(
		r.HTTPRequestsTotal,
		r.HTTPRequestDuration,
		r.GRPCRequestsTotal,
		r.GRPCRequestDuration,
		r.RateLimitRejectedTotal,
		r.CircuitBreakerRejectedTotal,
		r.CircuitBreakerState,
		r.AIRequestsTotal,
		r.AIRequestDuration,
	)

	return r
}

// Handler returns an HTTP handler exposing Prometheus metrics.
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.promRegistry, promhttp.HandlerOpts{})
}

// RecordRateLimitRejected increments rate limit rejection counter.
func (r *Registry) RecordRateLimitRejected(protocol string) {
	if r == nil {
		return
	}
	r.RateLimitRejectedTotal.WithLabelValues(protocol).Inc()
}

// RecordCircuitBreakerRejected increments circuit breaker rejection counter.
func (r *Registry) RecordCircuitBreakerRejected(protocol string) {
	if r == nil {
		return
	}
	r.CircuitBreakerRejectedTotal.WithLabelValues(protocol).Inc()
}

// SetCircuitBreakerState updates the circuit breaker state gauge.
func (r *Registry) SetCircuitBreakerState(name string, state gobreaker.State) {
	if r == nil {
		return
	}
	r.CircuitBreakerState.WithLabelValues(name).Set(float64(breakerStateValue(state)))
}

func breakerStateValue(state gobreaker.State) int {
	switch state {
	case gobreaker.StateClosed:
		return 0
	case gobreaker.StateHalfOpen:
		return 1
	case gobreaker.StateOpen:
		return 2
	default:
		return -1
	}
}

func statusLabel(code int) string {
	return strconv.Itoa(code)
}
