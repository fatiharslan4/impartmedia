package tags

import (
	"encoding/json"
	"fmt"
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

func TestHiveCompare(t *testing.T) {

	income := TagComparison{
		TagID:        Income().ID,
		SortOrder:    2,
		DisplayScope: "year",
		Percentiles: Percentiles{
			{
				Percent:   25,
				HighValue: 40000,
				Fill:      false,
			},
			{
				Percent:   50,
				HighValue: 75000,
				Fill:      false,
			},
			{
				Percent:   75,
				HighValue: 100000,
				Fill:      false,
			},
			{
				Percent:   100,
				HighValue: 125000,
				Fill:      false,
			},
		},
	}

	savings := TagComparison{
		TagID:        EmergencySavings().ID,
		SortOrder:    2,
		DisplayScope: "household",
		Percentiles: Percentiles{
			{
				Percent:   25,
				HighValue: 40000,
				Fill:      false,
			},
			{
				Percent:   50,
				HighValue: 75000,
				Fill:      false,
			},
			{
				Percent:   75,
				HighValue: 100000,
				Fill:      false,
			},
			{
				Percent:   100,
				HighValue: 125000,
				Fill:      false,
			},
		},
	}

	education := TagComparison{
		TagID:        EducationSavings().ID,
		SortOrder:    2,
		DisplayScope: "year",
		Percentiles: Percentiles{
			{
				Percent:   25,
				HighValue: 40000,
				Fill:      false,
			},
			{
				Percent:   50,
				HighValue: 75000,
				Fill:      false,
			},
			{
				Percent:   75,
				HighValue: 100000,
				Fill:      false,
			},
			{
				Percent:   100,
				HighValue: 125000,
				Fill:      false,
			},
		},
	}

	retirement := TagComparison{
		TagID:        RetirementSavings().ID,
		SortOrder:    2,
		DisplayScope: "year",
		Percentiles: Percentiles{
			{
				Percent:   25,
				HighValue: 40000,
				Fill:      false,
			},
			{
				Percent:   50,
				HighValue: 75000,
				Fill:      false,
			},
			{
				Percent:   75,
				HighValue: 100000,
				Fill:      false,
			},
			{
				Percent:   100,
				HighValue: 125000,
				Fill:      false,
			},
		},
	}

	lifeInsurance := TagComparison{
		TagID:        LifeInsuranceCoverage().ID,
		SortOrder:    2,
		DisplayScope: "year",
		Percentiles: Percentiles{
			{
				Percent:   25,
				HighValue: 40000,
				Fill:      false,
			},
			{
				Percent:   50,
				HighValue: 75000,
				Fill:      false,
			},
			{
				Percent:   75,
				HighValue: 100000,
				Fill:      false,
			},
			{
				Percent:   100,
				HighValue: 125000,
				Fill:      false,
			},
		},
	}

	netWorth := TagComparison{
		TagID:        NetWorth().ID,
		SortOrder:    2,
		DisplayScope: "year",
		Percentiles: Percentiles{
			{
				Percent:   25,
				HighValue: 40000,
				Fill:      false,
			},
			{
				Percent:   50,
				HighValue: 75000,
				Fill:      false,
			},
			{
				Percent:   75,
				HighValue: 100000,
				Fill:      false,
			},
			{
				Percent:   100,
				HighValue: 125000,
				Fill:      false,
			},
		},
	}

	ts := AvailableTags()
	b, _ := json.MarshalIndent(&ts, "", "\t")
	fmt.Println(string(b))

	c := TagComparisons{
		income, savings, education, retirement, lifeInsurance, netWorth,
	}

	b, _ = json.MarshalIndent(&c, "", "\t")
	fmt.Println(string(b))
}
