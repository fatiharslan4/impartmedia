package config

import (
	"fmt"
	"net/http"

	"github.com/xeipuuv/gojsonschema"

	"github.com/kelseyhightower/envconfig"
	"github.com/ory/graceful"
)

type Environment string

const (
	Local         Environment = "local"
	Development   Environment = "dev"
	IOS           Environment = "IOS-dev"
	Preproduction Environment = "preprod"
	Production    Environment = "prod"
)

var validEnvironments = []Environment{Local, Development, IOS, Preproduction, Production}

// all fields read from the environment, and prefixed with IMPART_
type Impart struct {
	Env                Environment `split_words:"true" default:"dev"`
	Debug              bool        `split_words:"true" default:"false"`
	Port               int         `split_words:"true" default:"8080"`
	APIKey             string      `split_words:"true" default:"38c31c4a79c04fd102e105f23a7cdcf832e40ad1b1a526ba82da9fe1f86aa5aab288a3f1a85f5edf39478d65c05c6f1328c82de7e1677ca31a4392ab"` //default is dev api key, generated via "openssl rand -hex 60"
	Region             string      `split_words:"true" default:"us-east-2"`
	DynamoEndpoint     string      `split_words:"true" default:"http://localhost:8000"`
	IOSNotificationARN string      `split_words:"true" default:""`
	ProfileSchemaPath  string      `split_words:"true" default:"./Profile.json"`
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
		Addr: fmt.Sprintf(":%v", ic.Port),
	})
	return server
}

func (ic Impart) GetProfileSchemaValidator() (gojsonschema.JSONLoader, error) {
	v := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", ic.ProfileSchemaPath))
	_, err := v.LoadJSON()
	return v, err
}
