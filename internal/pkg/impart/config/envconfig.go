package config

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/kelseyhightower/envconfig"
	"github.com/ory/graceful"
	"github.com/xeipuuv/gojsonschema"
)

type Environment string

func (e Environment) String() string {
	return string(e)
}

const (
	Local         Environment = "local"
	Development   Environment = "dev"
	IOS           Environment = "iosdev"
	Preproduction Environment = "preprod"
	Production    Environment = "prod"
)

var validEnvironments = []Environment{Local, Development, IOS, Preproduction, Production}

// media configuraitons
type Media struct {
	Storage      string `split_words:"true"`
	MediaPath    string `split_words:"true"`
	BucketName   string `split_words:"true"`
	BucketRegion string `split_words:"true"`
}

// all fields read from the environment, and prefixed with IMPART_
type Impart struct {
	Env    Environment `split_words:"true" default:"dev"`
	Debug  bool        `split_words:"true" default:"false"`
	Port   int         `split_words:"true" default:"8080"`
	APIKey string      `split_words:"true" default:"38c31c4a79c04fd102e105f23a7cdcf832e40ad1b1a526ba82da9fe1f86aa5aab288a3f1a85f5edf39478d65c05c6f1328c82de7e1677ca31a4392ab"` //default is dev api key, generated via "openssl rand -hex 60"
	Region string      `split_words:"true" default:"us-east-2"`
	//DynamoEndpoint     string      `split_words:"true" default:"http://localhost:8000"`
	IOSNotificationARN string `split_words:"true" default:""`
	ProfileSchemaPath  string `split_words:"true" default:"./schemas/json/Profile.json"`
	MigrationsPath     string `split_word:"true" default:"schemas/migrations"`

	DBHost     string `split_words:"true" default:"localhost"`
	DBPort     int    `split_words:"true" default:"3306"`
	DBName     string `split_words:"true" default:"impart"`
	DBUsername string `split_words:"true"`
	DBPassword string `split_words:"true"`

	DBMigrationUsername string            `split_words:"true"`
	DBMigrationPassword string            `split_words:"true"`
	SentryDSN           string            `split_words:"true"`
	Media               map[string]string `split_words:"true"`
}

func GetImpart() (*Impart, error) {
	var cfg Impart
	if err := envconfig.Process("impart", &cfg); err != nil {
		return nil, err
	}
	isValidEnvironment := false
	for _, e := range validEnvironments {
		if e == cfg.Env {
			isValidEnvironment = true
		}
	}
	if !isValidEnvironment {
		return nil, fmt.Errorf("invalid environment specified %s", cfg.Env)
	}

	return &cfg, nil
}

func (ic Impart) GetHttpServer() *http.Server {
	server := graceful.WithDefaults(&http.Server{
		Addr:        fmt.Sprintf(":%v", ic.Port),
		ReadTimeout: time.Second * 20,
	})
	return server
}

func (ic Impart) GetProfileSchemaValidator() (gojsonschema.JSONLoader, error) {
	v := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", ic.ProfileSchemaPath))
	_, err := v.LoadJSON()
	return v, err
}

type ZapBoilWriter struct {
	*zap.Logger
}

func (l *ZapBoilWriter) Write(p []byte) (n int, err error) {
	l.Debug(string(p))
	return len(p), nil
}
