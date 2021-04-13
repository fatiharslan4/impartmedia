package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/impartwealthapp/backend/pkg/client"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
	"time"
)

var zapConfig = &zap.Config{
	Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
	Development: false,
	Sampling: &zap.SamplingConfig{
		Initial:    1000,
		Thereafter: 1000,
	},
	Encoding: "console",
	EncoderConfig: zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	},
	OutputPaths:       []string{"stderr"},
	ErrorOutputPaths:  []string{"stderr"},
	DisableStacktrace: true,
	DisableCaller:     true,
}

var apiEnvironmentKeys = map[string]string{
	"local":   "38c31c4a79c04fd102e105f23a7cdcf832e40ad1b1a526ba82da9fe1f86aa5aab288a3f1a85f5edf39478d65c05c6f1328c82de7e1677ca31a4392ab",
	"dev":     "AAs8wLBVd41EEO7Qws25ocutQAjuzwz5MM1nNNLa",
	"iosdev":  "yCwm0JHpIR49GLTG8pqnd6lmTo10Cw2b5gr9qGNM",
	"preprod": "K39z2qMKV959GdI7sWpczbnhyiw4Zno6RCVXh233",
	"prod":    "I1TuBFDPdh5vRYdqqIRDn7OqITyyPIQO3SQnemuS",
}

type Config struct {
	Environment       string `short:"E" required:"true" help:"environment to execute against.  Must be one of local, dev, iosdev, preprod or prod"`
	Debug             bool   `short:"D" help:"Enable debug mode"`
	Auth0ClientID     string `required:"true" env:"AUTH0_CLIENT_ID" help:"This value is The Client ID from the Auth0 \"Auth0 Management API\" application. Ideally this is set via the environment variable"`
	Auth0ClientSecret string `required:"true" env:"AUTH0_CLIENT_SECRET" help:"This value is The Client Secret from the Auth0 \"Auth0 Management API\" application. Ideally this is set via the environment variable"`
}

func (c *Config) NewImpartManagementClient(logger *zap.Logger) client.ImpartMangementClient {
	return client.NewManagement(c.Environment, apiEnvironmentKeys[c.Environment], client.Auth0Credentials{
		ClientID:     c.Auth0ClientID,
		ClientSecret: c.Auth0ClientSecret,
	}, logger)
}

func (c *Config) Validate() error {
	if _, validEnv := apiEnvironmentKeys[c.Environment]; !validEnv {
		allowedEnvironments := make([]string, len(apiEnvironmentKeys), len(apiEnvironmentKeys))
		i := 0
		for k, _ := range apiEnvironmentKeys {
			allowedEnvironments[i] = k
			i++
		}
		return fmt.Errorf("invalid environment %s, must be one of %s", c.Environment, strings.Join(allowedEnvironments, ", "))
	}

	return nil
}

type CLI struct {
	Config
	CreateUser CreateUserCmd `cmd:"true" help:"Creates an Auth0 User and Impart Wealth in the environment set globally; returns the impartWealthID."`
	DeleteUser DeleteUserCmd `cmd:"true" help:"Deletes an Auth0 User by their email or impartWealthId. Requires an impart admin account."`
}

func main() {
	start := time.Now()
	cli := CLI{
		Config: Config{},
	}
	logger, _ := zapConfig.Build()
	defer logger.Sync()

	ctx := kong.Parse(&cli,
		kong.Name("impart"),
		kong.Description("a CLI for managing the impart wealth backend in various environments"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": "0.0.1",
		})
	if cli.Config.Debug {
		zapConfig.Development = true
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		zapConfig.EncoderConfig.CallerKey = "C"
		logger, _ = zapConfig.Build()
	}
	ctx.Bind(logger)
	ctx.BindTo(cli.Config.NewImpartManagementClient(logger), (*client.ImpartMangementClient)(nil))
	logger.Info("ImpartWealth Management Client Start", zap.String("environment", cli.Config.Environment))
	err := ctx.Run(&cli.Config)
	if err != nil {
		logger.Fatal("error executing command", zap.Error(err), zap.Duration("duration", time.Since(start)))
	} else {
		logger.Info("success", zap.Duration("duration", time.Since(start)))
	}
}

type CreateUserCmd struct {
	Email         string `short:"e" required:""`
	Password      string `short:"p" required:""`
	ScreenName    string `short:"s" optional:"" help:"Create the user under an explicit screenName, otherwise it will be set to the email"`
	AdminUsername string `short:"U" required:"" help:"impart wealth administrator account username"`
	AdminPassword string `short:"P" required:"" help:"impart wealth administrator account password"`
}

func (c *CreateUserCmd) Run(cfg *Config, managementclient client.ImpartMangementClient, logger *zap.Logger) error {
	fmt.Printf("creating a new user in %s with email %s\n", cfg.Environment, c.Email)

	resp, err := managementclient.CreateUser(client.CreateUserRequest{
		Environment:   cfg.Environment,
		Email:         c.Email,
		Password:      c.Password,
		ScreenName:    c.ScreenName,
		AdminUsername: c.AdminUsername,
		AdminPassword: c.AdminPassword,
	})
	if err != nil {
		return err
	}
	logger.Info("Impart Wealth Account Created", zap.Any("createdUser", resp))
	return nil
}

type DeleteUserCmd struct {
	Email          string `short:"e" help:"the email address of the auth0 user to delete; ignored if impartWealthId is set"`
	ImpartWealthID string `name:"id" short:"i" help:"the impartWealthId of the user to delete; do not pass both"`
	AdminUsername  string `short:"U" required:"" help:"impart wealth administrator account username"`
	AdminPassword  string `short:"P" required:"" help:"impart wealth administrator account password"`
}

func (c *DeleteUserCmd) Run(cfg *Config, managementClient client.ImpartMangementClient, logger *zap.Logger) error {
	logger.Info("deleting all users matching input args", zap.String("impartWealthID", c.ImpartWealthID), zap.String("email", c.Email))

	_, err := managementClient.DeleteUser(client.DeleteUserRequest{
		Environment:    cfg.Environment,
		Email:          c.Email,
		ImpartWealthID: c.ImpartWealthID,
		AdminUsername:  c.AdminUsername,
		AdminPassword:  c.AdminPassword,
	})
	if err != nil {
		return err
	}
	return nil
}
