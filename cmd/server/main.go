package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/kelseyhightower/envconfig"
	"github.com/ory/graceful"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	var cfg config.Impart
	if err = envconfig.Process("impart", &cfg); err != nil {
		log.Fatal(err.Error())
	}
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	// Example ping request.
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
	})

	// Example when panic happen.
	r.GET("/panic", func(c *gin.Context) {
		panic("An unexpected error happen!")
	})

	server := graceful.WithDefaults(&http.Server{
		Handler: r,
		Addr:    ":8080",
	})

	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		logger.Fatal("error serving", zap.Error(err))
	}
	logger.Info("done serving")
}
