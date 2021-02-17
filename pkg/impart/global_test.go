package impart

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupExecutionContext(t *testing.T) {

	os.Setenv("REGION", "myregion")

	type addlConfig struct {
		SomeMode bool `split_words:"true" default:"false"`
	}
	os.Setenv("SOME_MODE", "true")

	cfg := addlConfig{SomeMode: false}

	execCtx := SetupExecutionContext("unit", &cfg)

	assert.NotNil(t, execCtx)
	assert.NotNil(t, execCtx.Config)
	assert.NotNil(t, execCtx.Logger)
	assert.NotNil(t, cfg)
	assert.True(t, cfg.SomeMode)
	assert.Equal(t, "myregion", execCtx.Region)
}
