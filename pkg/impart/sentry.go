package impart

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitSentryLogger(cfg *config.Impart, log *zap.Logger, debug bool) (logger *zap.Logger, err error) {
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
