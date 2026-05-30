// Package http provides Gin-based HTTP server with HTTP/1, HTTP/2, and HTTP/3 support.
//
// Rate-limit and breaker middleware live in the http/middleware subpackage.
// app.Run mounts L1 GlobalRateLimit and BreakerCheck; L2/L3 are registered
// in business infrastructure/registerHTTP.
package http
