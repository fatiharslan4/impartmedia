package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProfile_EqualsIgnoreTimes(t *testing.T) {
	p := RandomProfile()
	p2 := p.Copy()
	p2.UpdatedDate = time.Now()
	p2.CreatedDate = time.Now()
	p2.Attributes.UpdatedDate = time.Now()
	p2.Attributes.Address.UpdatedDate = time.Now()
	p2.SurveyResponses.ImportTimestamp = time.Now()

	assert.False(t, p.Equals(p2))
	assert.NotEqual(t, p, p2)
	assert.True(t, p.EqualsIgnoreTimes(p2))
}
