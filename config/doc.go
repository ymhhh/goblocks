// Package config loads application configuration from YAML with env overrides.
//
// resilience.rate_limit supports layered limits: global (L1), user (L2), routes (L3),
// and backend selection (memory | redis). Legacy top-level rps/burst map to global.
// See docs/configuration.md for the full schema.
package config
