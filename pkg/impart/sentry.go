package impart

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/sentryCore"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitSentryLogger(cfg *config.Impart, log *zap.Logger, debug bool) (logger *zap.Logger, err error) {
	logger = log
	if cfg.SentryDSN != "" {
		if debug {
			logger, err = zap.NewDevelopment()
		} else {
			logger, err = zap.NewProduction()
		}

		logger = ModifyToSentryLogger(logger, cfg.SentryDSN, cfg.Env.String())

	}
	return
}

func ModifyToSentryLogger(log *zap.Logger, DSN string, env string) *zap.Logger {
	cfg := sentryCore.Configuration{
		Level: zapcore.ErrorLevel, //when to send message to sentry
		Tags: map[string]string{
			"component": "system",
		},
		DisableStacktrace: true,
	}
	core, err := sentryCore.NewCore(
		cfg,
		sentryCore.NewSentryClientFromDSN(DSN),
		sentryCore.SentryEventConfig{
			ServerName:  fmt.Sprintf("impart-%s", env),
			Platform:    "Golang",
			Environment: env,
		},
	)

	//in case of err it will return noop core. so we can safely attach it
	if err != nil {
		log.Warn("failed to init zap", zap.Error(err))
	}
	return AttachCoreToLogger(core, log)
}

func AttachCoreToLogger(sentryCore zapcore.Core, l *zap.Logger) *zap.Logger {
	return l.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, sentryCore)
	}))
}

func InitSentryLoggerBackup(cfg *config.Impart, log *zap.Logger, debug bool) (logger *zap.Logger, err error) {
	logger = log

	var ZapHook = zap.Hooks(func(entry zapcore.Entry) error {
		if entry.Level == zapcore.ErrorLevel || entry.Level == zapcore.FatalLevel {
			defer sentry.Flush(2 * time.Second)

			messageString := fmt.Sprintf("%s, Line No: %d :: %s",
				entry.Caller.TrimmedPath(), entry.Caller.Line, entry.Message,
			)
			level := sentry.LevelError
			sentry.CaptureEvent(&sentry.Event{
				ServerName: fmt.Sprintf("impart-%s", cfg.Env),
				Level:      level,
				Message:    fmt.Sprintf("%s \n\n\n %s", messageString, entry.Stack),
				Exception: []sentry.Exception{
					{
						Type:  fmt.Sprintf("impart-%s", cfg.Env),
						Value: entry.Message,
					},
				},
			})
		}
		return nil
	})

	if cfg.SentryDSN != "" {
		if err = sentry.Init(sentry.ClientOptions{
			Dsn:         cfg.SentryDSN,
			Environment: cfg.Env.String(),
		}); err != nil {
			return log, err
		}

		if debug {
			logger, err = zap.NewDevelopment(ZapHook)
		} else {
			logger, err = zap.NewProduction(ZapHook)
		}
	} else {
		logger.Info("Sentry logger is not configured")
	}
	return
}
