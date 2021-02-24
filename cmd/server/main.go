package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/impartwealthapp/backend/pkg/tags"

	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/auth"
	hivedata "github.com/impartwealthapp/backend/pkg/data/hive"
	profiledata "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/hive"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/profile"
	"github.com/ory/graceful"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.GetImpart()
	if err != nil {
		logger.Fatal("error parsing config", zap.Error(err))
	}
	if cfg == nil {
		logger.Fatal("nil config")
		return
	}

	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
		logger, _ = zap.NewDevelopment()
		if cfg.Env == config.Local || cfg.Env == config.Development {
			logger.Debug("config startup", zap.Any("config", cfg))
		}

	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	defer logger.Sync()

	services := setupServices(cfg, logger)

	r := gin.New()
	r.RedirectTrailingSlash = true
	r.Use(ginzap.RecoveryWithZap(logger, true))      // panics don't stop server
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true)) // logs all requests

	r.NoRoute(noRouteFunc)
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	v1 := r.Group(fmt.Sprintf("%s/v1", cfg.Env))
	v1.Use(services.Auth.APIKeyHandler())               //x-api-key is present on all requests
	v1.Use(services.Auth.RequestAuthorizationHandler()) //ensure request has valid JWT
	v1.GET("/tags", func(ctx *gin.Context) { ctx.JSON(http.StatusOK, tags.AvailableTags()) })

	hive.SetupRoutes(v1, services.HiveData, services.Hive, logger)
	profile.SetupRoutes(v1, services.ProfileData, services.Profile, logger)

	server := cfg.GetHttpServer()
	server.Handler = r
	logger.Info("Impart backend started.", zap.Int("port", cfg.Port), zap.String("env", string(cfg.Env)))
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		logger.Fatal("error serving", zap.Error(err))
	}
	logger.Info("done serving")
}

func noRouteFunc(ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, impart.NewError(impart.ErrBadRequest, fmt.Sprintf("unable to find a valid http route for %s", ctx.Request.RequestURI)))
}

type Services struct {
	ProfileData   profiledata.Store
	Profile       profile.Service
	Hive          hive.Service
	HiveData      hivedata.Hives
	Auth          auth.Service
	Notifications impart.NotificationService
}

func setupServices(cfg *config.Impart, logger *zap.Logger) *Services {
	var err error
	svcs := &Services{}
	svcs.ProfileData, err = profiledata.New(cfg.Region, cfg.DynamoEndpoint, string(cfg.Env), logger.Sugar())
	if err != nil {
		logger.Fatal("err creating profile data service", zap.Error(err))
	}
	svcs.HiveData, err = hivedata.NewHiveData(cfg.Region, cfg.DynamoEndpoint, string(cfg.Env), logger)

	svcs.Auth, err = auth.NewAuthService(cfg, svcs.ProfileData, logger)
	if err != nil {
		logger.Fatal("err creating auth service", zap.Error(err))
	}

	if strings.Contains(cfg.DynamoEndpoint, "localhost") || strings.Contains(cfg.DynamoEndpoint, "127.0.0.1") {
		svcs.Notifications = impart.NewNoopNotificationService()
	} else {
		svcs.Notifications = impart.NewImpartNotificationService(string(cfg.Env), cfg.Region, cfg.IOSNotificationARN, logger)
	}

	profileValidator, err := cfg.GetProfileSchemaValidator()
	if err != nil {
		logger.Fatal("err creating profile schema validator", zap.Error(err))
	}

	svcs.Profile = profile.New(logger.Sugar(), svcs.ProfileData, svcs.Notifications, profileValidator, string(cfg.Env))

	svcs.Hive = hive.New(cfg.Region, cfg.DynamoEndpoint, string(cfg.Env), cfg.IOSNotificationARN, logger)

	return svcs
}
