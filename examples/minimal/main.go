package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ymhhh/goblocks/app"
	"github.com/ymhhh/goblocks/config"
	"github.com/ymhhh/goblocks/resilience"
)

func main() {
	cfgPath := "config/config.yaml"
	if v := os.Getenv("CONFIG_PATH"); v != "" {
		cfgPath = v
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		panic(err)
	}
	cfg.Server.GRPC.Enabled = false

	application := app.New(cfg).WithHTTP(func(engine *gin.Engine, _ *resilience.Policy) {
		engine.GET("/hello", func(c *gin.Context) {
			c.String(http.StatusOK, "hello from goblocks")
		})
	})

	if err := application.Run(context.Background()); err != nil {
		panic(err)
	}
}
