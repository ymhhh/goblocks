package resilience

import (
	"testing"

	"github.com/ymhhh/goblocks/config"
)

func TestBuildRateLimitRules(t *testing.T) {
	cfg := config.RateLimitConfig{
		Global: config.RateLimitTierConfig{RPS: 50, Burst: 100},
		User: config.UserRateLimitConfig{
			Enabled:    true,
			DefaultRPS: 10,
			Burst:      20,
		},
		Routes: []config.RouteRateLimitConfig{
			{Method: "post", Path: "/ai/chat", RPS: 5, Burst: 5},
		},
	}

	global, user, userEnabled, routes := BuildRateLimitRules(cfg)
	if global.RPS != 50 || global.Burst != 100 {
		t.Fatalf("global rule: %+v", global)
	}
	if !userEnabled || user.RPS != 10 || user.Burst != 20 {
		t.Fatalf("user rule: enabled=%v rule=%+v", userEnabled, user)
	}
	if len(routes) != 1 {
		t.Fatalf("routes: %+v", routes)
	}
	rule, ok := routes["POST:/ai/chat"]
	if !ok || rule.RPS != 5 || rule.Burst != 5 {
		t.Fatalf("route rule: ok=%v rule=%+v", ok, rule)
	}
}
