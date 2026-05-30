// Package metrics provides Prometheus observability for goblocks services.
//
// Rate-limit rejections are recorded via RecordRateLimitRejected(protocol, scope)
// where scope is global, user, or route (L1/L2/L3).
package metrics
