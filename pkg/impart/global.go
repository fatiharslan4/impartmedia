package impart

import (
	"net"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

// ExecutionContext provides the re-usable execution context within a lambda function.
type ExecutionContext struct {
	Warm   bool
	Debug  bool
	Logger *zap.Logger
	Stage  string
	*Config
}

// Config represents the environment configuration
type Config struct {
	Region string `split_words:"true"`
}

// SetupExecutionContext parses the base ImpartLambda configuration
// and takes any additional configurations to parse.
// this function is mean to be called on startup, and will panic on any errors
func SetupExecutionContext(stage string, additionalConfig interface{}) ExecutionContext {
	cfg := &Config{}

	if err := envconfig.Process("", cfg); err != nil {
		panic("unable to load env variables" + err.Error())
	}

	if additionalConfig != nil {
		v := reflect.ValueOf(additionalConfig)

		if v.Kind() != reflect.Ptr {
			panic("Expected a pointer to a configuration value")
		}

		if err := envconfig.Process("", additionalConfig); err != nil {
			panic("unable to load env variables for additional config: " + err.Error())
		}
	}

	logger, err := zap.NewProduction()

	if err != nil {
		panic("unable to instantiate zap production logger" + err.Error())
	}

	execContext := ExecutionContext{
		Warm:   true,
		Debug:  false,
		Logger: logger.Named(stage),
		Stage:  stage,
		Config: cfg,
	}

	return execContext
}

func AuthIdFromRequest(request events.APIGatewayProxyRequest) string {
	return AuthIdFromContext(request.RequestContext)
}

func AuthIdFromContext(ctx events.APIGatewayProxyRequestContext) string {
	authId, ok := ctx.Authorizer["authenticationId"]
	if !ok {
		return ""
	}
	return authId.(string)
}

func NewDevelopmentLogger(stage string) *zap.Logger {
	l, _ := zap.NewDevelopment()
	return l.Named(stage)
}

func NewProductionLogger(stage string) *zap.Logger {
	l, _ := zap.NewProduction()
	return l.Named(stage)
}

func CurrentUTC() time.Time {
	return time.Now().In(time.UTC)
}

func ImpartHttpClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       20 * time.Second,
		TLSHandshakeTimeout:   2 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 5,
	}
	return &http.Client{
		Transport: transport,
	}
}
