package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

func registerHealthRoutes(engine *gin.Engine, a *App) {
	if engine == nil || a == nil || !a.cfg.Server.HTTP.Health.Enabled {
		return
	}

	liveness := a.cfg.Server.HTTP.Health.LivenessPath
	if liveness == "" {
		liveness = "/health"
	}
	readiness := a.cfg.Server.HTTP.Health.ReadinessPath
	if readiness == "" {
		readiness = "/ready"
	}

	engine.GET(liveness, func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	engine.GET(readiness, func(c *gin.Context) {
		if !a.ready() {
			c.String(http.StatusServiceUnavailable, "not ready")
			return
		}
		c.String(http.StatusOK, "ok")
	})
}

func (a *App) ready() bool {
	if a.policy != nil && a.policy.Breaker != nil {
		if a.policy.Breaker.State() == gobreaker.StateOpen {
			return false
		}
	}
	if a.cfg.Server.GRPC.Enabled && a.grpcRegister != nil && a.grpcServer == nil {
		return false
	}
	return true
}
