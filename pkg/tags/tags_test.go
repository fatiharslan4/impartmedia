package tags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromID(t *testing.T) {

	tag, err := FromID(0)
	assert.Error(t, err)
	assert.Equal(t, 0, tag.ID)

	tag, err = FromID(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, tag.ID)

	tag, err = FromID(2)
	assert.NoError(t, err)
	assert.Equal(t, 2, tag.ID)

	tag, err = FromID(3)
	assert.NoError(t, err)
	assert.Equal(t, 3, tag.ID)

	tag, err = FromID(4)
	assert.NoError(t, err)
	assert.Equal(t, 4, tag.ID)

	tag, err = FromID(5)
	assert.NoError(t, err)
	assert.Equal(t, 5, tag.ID)

	tag, err = FromID(6)
	assert.NoError(t, err)
	assert.Equal(t, 6, tag.ID)

	tag, err = FromID(7)
	assert.NoError(t, err)
	assert.Equal(t, 7, tag.ID)

}

func TestAvailableTags(t *testing.T) {

	allTags := AvailableTags()
	assert.Len(t, allTags, 7)
	assert.Equal(t, allTags[0], Income())
	assert.Equal(t, allTags[1], EmergencySavings())
	assert.Equal(t, allTags[2], EducationSavings())
	assert.Equal(t, allTags[3], RetirementSavings())
	assert.Equal(t, allTags[4], LifeInsuranceCoverage())
	assert.Equal(t, allTags[5], NetWorth())
	assert.Equal(t, allTags[6], Other())
}
