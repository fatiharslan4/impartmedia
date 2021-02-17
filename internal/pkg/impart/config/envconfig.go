package config

// all fields read from the environment, and prefixed with IMPART_
type Impart struct {
	Debug bool `default:"false"`
	Port  int  `default:"8080"`
}
