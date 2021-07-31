package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/impartwealthapp/backend/pkg/data/migrater"
	"github.com/impartwealthapp/backend/pkg/media"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/sqlboiler/v4/boil"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/auth"
	hivedata "github.com/impartwealthapp/backend/pkg/data/hive"
	profiledata "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/hive"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/profile"
	"github.com/impartwealthapp/backend/pkg/tags"
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
		//boil.DebugMode = true
		boil.WithDebugWriter(context.TODO(), &config.ZapBoilWriter{Logger: logger})
		logger, _ = zap.NewDevelopment()
		if cfg.Env == config.Local || cfg.Env == config.Development {
			logger.Debug("config startup", zap.Any("config", *cfg))
		}
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	//init the sentry logger ,either debug
	logger, err = impart.InitSentryLogger(cfg, logger, cfg.Debug)
	if err != nil {
		logger.Error("error on sentry init", zap.Any("error", err))
	}

	migrationDB, err := cfg.GetMigrationDBConnection()
	if err != nil {
		logger.Fatal("unable to connect to DB", zap.Error(err))
	}

	//Trap sigterm during migraitons
	migrationsDoneChan := make(chan bool)
	shutdownMigrationsChan := make(chan bool)
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		select {
		case <-sigc:
			logger.Info("received a shutdown request during migrations, sending shutdown signal")
			shutdownMigrationsChan <- true
		case <-migrationsDoneChan:
			logger.Info("migrations complete, no longer waiting for sig int")
			return
		}
	}()

	err = migrater.RunMigrationsUp(migrationDB, cfg.MigrationsPath, logger, shutdownMigrationsChan)
	if err != nil {
		logger.Fatal("error running migrations", zap.Error(err))
	}
	migrationsDoneChan <- true
	if err := migrationDB.Close(); err != nil {
		logger.Fatal("error closing migrations DB connection", zap.Error(err))
	}

	boil.SetLocation(time.UTC)
	db, err := cfg.GetDBConnection()
	if err != nil {
		logger.Fatal("unable to connect to DB", zap.Error(err))
	}
	defer db.Close()
	defer logger.Sync()

	if err := migrater.BootStrapAdminUsers(db, cfg.Env, logger); err != nil {
		logger.Fatal("unable to bootstrap user", zap.Error(err))
	}

	if err := migrater.BootStrapTopicHive(db, cfg.Env, logger); err != nil {
		logger.Fatal("unable to bootstrap user", zap.Error(err))
	}

	// initiate global profanity detector
	impart.InitProfanityDetector(db, logger)

	services := setupServices(cfg, db, logger)

	r := gin.New()
	r.Use(CORS)
	r.RedirectTrailingSlash = true
	r.Use(ginzap.RecoveryWithZap(logger, true))      // panics don't stop server
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true)) // logs all requests

	r.NoRoute(noRouteFunc)
	r.GET("/ping", func(ctx *gin.Context) {
		_, err := dbmodels.Pings(dbmodels.PingWhere.Ok.EQ(true)).One(ctx, db)
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
		}
		ctx.String(http.StatusOK, "pong")
	})
	var v1Route string
	if cfg.Env == config.Production || cfg.Env == config.Local {
		v1Route = "v1"
	} else {
		v1Route = fmt.Sprintf("%s/v1", cfg.Env)
	}

	v1 := r.Group(v1Route)

	v1.Use(services.Auth.APIKeyHandler())               //x-api-key is present on all requests
	v1.Use(services.Auth.RequestAuthorizationHandler()) //ensure request has valid JWT
	v1.Use(services.Auth.DeviceIdentificationHandler()) //context for device identification
	v1.GET("/tags", func(ctx *gin.Context) { ctx.JSON(http.StatusOK, tags.AvailableTags()) })

	hive.SetupRoutes(v1, db, services.HiveData, services.Hive, logger)
	profile.SetupRoutes(v1, services.ProfileData, services.Profile, logger, services.Notifications)

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
	MediaStorage  media.StorageConfigurations
}

func setupServices(cfg *config.Impart, db *sql.DB, logger *zap.Logger) *Services {
	var err error
	svcs := &Services{}
	svcs.ProfileData = profiledata.NewMySQLStore(db, logger)
	svcs.HiveData = hivedata.NewHiveService(db, logger)

	svcs.Auth, err = auth.NewAuthService(cfg, svcs.ProfileData, logger)
	if err != nil {
		logger.Fatal("err creating auth service", zap.Error(err))
	}

	if cfg.Env == config.Local {
		svcs.Notifications = impart.NewNoopNotificationService()
	} else {
		svcs.Notifications = impart.NewImpartNotificationService(db, string(cfg.Env), cfg.Region, cfg.IOSNotificationARN, logger)
	}

	profileValidator, err := cfg.GetProfileSchemaValidator()
	if err != nil {
		logger.Fatal("err creating profile schema validator", zap.Error(err))
	}

	svcs.Profile = profile.New(logger.Sugar(), db, svcs.ProfileData, svcs.Notifications, profileValidator, string(cfg.Env))

	svcs.MediaStorage = media.LoadMediaConfig(cfg)
	svcs.Hive = hive.New(cfg, db, logger, svcs.MediaStorage)
	return svcs
}

func CORS(c *gin.Context) {

	// First, we add the headers with need to enable CORS
	// Make sure to adjust these headers to your needs
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Content-Type", "application/json")
	// Second, we handle the OPTIONS problem
	if c.Request.Method != "OPTIONS" {

		c.Next()

	} else {

		// Everytime we receive an OPTIONS request,
		// we just return an HTTP 200 Status Code
		// Like this, Angular can now do the real
		// request using any other method than OPTIONS
		c.AbortWithStatus(http.StatusOK)
	}
}
