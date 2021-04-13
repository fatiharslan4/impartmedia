package hive

import (
	"encoding/json"
	"testing"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/tags"
	"github.com/stretchr/testify/assert"
)

func TestFillPercentiles(t *testing.T) {
	p := tags.Percentiles{
		{Percent: 25, HighValue: 100},
		{Percent: 50, HighValue: 1000},
		{Percent: 75, HighValue: 10000},
		{Percent: 100, HighValue: 100000},
	}

	ourVal := 9000
	expectedHighPercent := 75
	expectedHighPosition := 3
	expectedHighPercentile := tags.Percentile{
		Percent: 75, HighValue: 10000, DisplayValue: impart.FriendlyFormatDollars(10000), Fill: true,
	}

	filledPercentiles, hv, hp := fillPercentiles(ourVal, p)

	assert.True(t, filledPercentiles[0].Fill)
	assert.True(t, filledPercentiles[1].Fill)
	assert.Equal(t, expectedHighPercentile, filledPercentiles[2])
	assert.Equal(t, expectedHighPercent, hv)
	assert.False(t, filledPercentiles[3].Fill)
	assert.Equal(t, expectedHighPosition, hp)

}

func TestCompare(t *testing.T) {
	p := tags.Percentiles{
		{Percent: 25, HighValue: 100},
		{Percent: 50, HighValue: 1000},
		{Percent: 75, HighValue: 10000},
		{Percent: 100, HighValue: 100000},
	}

	c := tags.TagComparison{
		TagID:        tags.IncomeID,
		SortOrder:    5,
		DisplayScope: "year",
		Percentiles:  p,
	}

	filledComparison := compare(9000, c)

	assert.Equal(t, 75, filledComparison.Percentile)
	assert.Equal(t, 9000, filledComparison.Value)
	assert.Equal(t, "$9K", filledComparison.DisplayValue)
	assert.Equal(t, "75th", filledComparison.DisplayPercentile)
	assert.Equal(t, "You are in the 3rd quartile of the Hive", filledComparison.DisplayDescription)

}

func TestCompareZeros(t *testing.T) {
	p := tags.Percentiles{
		{Percent: 25, HighValue: 100},
		{Percent: 50, HighValue: 1000},
		{Percent: 75, HighValue: 10000},
		{Percent: 100, HighValue: 100000},
	}

	c := tags.TagComparison{
		TagID:        tags.IncomeID,
		SortOrder:    5,
		DisplayScope: "year",
		Percentiles:  p,
	}

	filledComparison := compare(0, c)

	assert.Equal(t, 0, filledComparison.Percentile)
	assert.Equal(t, 0, filledComparison.Value)
	assert.Equal(t, "<$1K", filledComparison.DisplayValue)
	assert.Equal(t, "0th", filledComparison.DisplayPercentile)
	assert.Equal(t, "You are in the 0th quartile of the Hive", filledComparison.DisplayDescription)

}

func TestProfileCompare(t *testing.T) {
	jsonData := `[
  {
    "tagId": 1,
    "sortOrder": 1,
    "displayScope": "year",
    "percentiles": [
      {
        "percent": 25,
        "highValue": 40000
      },
      {
        "percent": 50,
        "highValue": 75000
      },
      {
        "percent": 75,
        "highValue": 100000
      },
      {
        "percent": 100,
        "highValue": 125000
      }
    ]
  },
  {
    "tagId": 2,
    "sortOrder": 2,
    "displayScope": "household",
    "percentiles": [
      {
        "percent": 25,
        "highValue": 40000
      },
      {
        "percent": 50,
        "highValue": 75000
      },
      {
        "percent": 75,
        "highValue": 100000
      },
      {
        "percent": 100,
        "highValue": 125000
      }
    ]
  },
  {
    "tagId": 3,
    "sortOrder": 5,
    "displayScope": "child",
    "percentiles": [
      {
        "percent": 25,
        "highValue": 40000
      },
      {
        "percent": 50,
        "highValue": 75000
      },
      {
        "percent": 75,
        "highValue": 100000
      },
      {
        "percent": 100,
        "highValue": 125000
      }
    ]
  },
  {
    "tagId": 4,
    "sortOrder": 4,
    "displayScope": "household",
    "percentiles": [
      {
        "percent": 25,
        "highValue": 40000
      },
      {
        "percent": 50,
        "highValue": 75000
      },
      {
        "percent": 75,
        "highValue": 100000
      },
      {
        "percent": 100,
        "highValue": 125000
      }
    ]
  },
  {
    "tagId": 5,
    "sortOrder": 3,
    "displayScope": "household",
    "percentiles": [
      {
        "percent": 25,
        "highValue": 40000
      },
      {
        "percent": 50,
        "highValue": 75000
      },
      {
        "percent": 75,
        "highValue": 100000
      },
      {
        "percent": 100,
        "highValue": 125000
      }
    ]
  },
  {
    "tagId": 7,
    "sortOrder": 6,
    "displayScope": "household",
    "percentiles": [
      {
        "percent": 25,
        "highValue": 40000
      },
      {
        "percent": 50,
        "highValue": 75000
      },
      {
        "percent": 75,
        "highValue": 100000
      },
      {
        "percent": 100,
        "highValue": 125000
      }
    ]
  }
]`

	var c tags.TagComparisons

	err := json.Unmarshal([]byte(jsonData), &c)
	assert.NoError(t, err)

	p := models.RandomProfile()

	outComparisons := profileCompare(p, c)

	outComparisons.SortAscending()
	assert.Equal(t, 6, len(outComparisons))

	//assert.Equal(t, tags.IncomeID, outComparisons[0].TagID)
	//assert.Equal(t, p.SurveyResponses.HouseholdIncomeAmount, outComparisons[0].Value)
	//assert.Equal(t, 50, outComparisons[0].Percentile)
	//assert.Equal(t, "You are in the 2nd quartile of the Hive", outComparisons[0].DisplayDescription)
	//
	//assert.Equal(t, tags.EmergencySavingsID, outComparisons[1].TagID)
	//assert.Equal(t, p.SurveyResponses.EmergencySavingsAmount, outComparisons[1].Value)
	//assert.Equal(t, 75, outComparisons[1].Percentile)
	//assert.Equal(t, "You are in the 3rd quartile of the Hive", outComparisons[1].DisplayDescription)
	//
	//assert.Equal(t, tags.RetirementSavingsID, outComparisons[3].TagID)
	//assert.Equal(t, p.SurveyResponses.RetirementSavingsAmount, outComparisons[3].Value)
	//assert.Equal(t, 100, outComparisons[3].Percentile)
	//assert.Equal(t, "You are in the 4th quartile of the Hive", outComparisons[3].DisplayDescription)
	//
	//assert.Equal(t, tags.LifeInsuranceCoverageID, outComparisons[2].TagID)
	//assert.Equal(t, p.SurveyResponses.LifeInsuranceAmount, outComparisons[2].Value)
	//assert.Equal(t, 100, outComparisons[2].Percentile)
	//
	//assert.Equal(t, tags.NetWorthID, outComparisons[5].TagID)
	//assert.Equal(t, p.SurveyResponses.NetWorthAmount, outComparisons[5].Value)
	//assert.Equal(t, 25, outComparisons[5].Percentile)
	//assert.Equal(t, "You are in the 1st quartile of the Hive", outComparisons[5].DisplayDescription)
	//
	//assert.Equal(t, tags.EducationSavingsID, outComparisons[4].TagID)
	//assert.Equal(t, p.SurveyResponses.EducationSavingsAmount, outComparisons[4].Value)
	//assert.Equal(t, 25, outComparisons[4].Percentile)

}
