package hive

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/tags"
)

func (s *service) HiveProfilePercentiles(ctx context.Context, hiveID uint64) (tags.TagComparisons, impart.Error) {
	var comparisons tags.TagComparisons
	ctxUser := impart.GetCtxUser(ctx)

	dbProfile, err := s.profileData.GetProfile(ctx, ctxUser.ImpartWealthID)
	if err != nil {
		s.logger.Error("error getting db profile", zap.Error(err))
		return tags.TagComparisons{}, impart.NewError(impart.ErrUnknown, "unable to get profile")
	}
	dbHive, err := s.hiveData.GetHive(ctx, hiveID)
	if err != nil {
		return comparisons, impart.NewError(err, fmt.Sprintf("error getting hive %v", hiveID))
	}

	hive, err := models.HiveFromDB(dbHive)
	if err != nil {
		s.logger.Error("error converting db hive", zap.Error(err))
		return tags.TagComparisons{}, impart.NewError(impart.ErrUnknown, "unable to build hive")
	}

	profile, err := models.ProfileFromDBModel(ctxUser, dbProfile)
	if err != nil {
		s.logger.Error("error getting db profile", zap.Error(err))
		return tags.TagComparisons{}, impart.NewError(impart.ErrUnknown, "unable to build profile")
	}

	return profileCompare(*profile, hive.TagComparisons), nil
}

func profileCompare(p models.Profile, comparisons tags.TagComparisons) tags.TagComparisons {
	for i, c := range comparisons {
		switch c.TagID {
		case tags.NetWorthID:
			c = compare(p.SurveyResponses.NetWorthAmount, c)
		case tags.EducationSavingsID:
			c = compare(p.SurveyResponses.EducationSavingsAmount, c)
		case tags.EmergencySavingsID:
			c = compare(p.SurveyResponses.EmergencySavingsAmount, c)
		case tags.RetirementSavingsID:
			c = compare(p.SurveyResponses.RetirementSavingsAmount, c)
		case tags.LifeInsuranceCoverageID:
			c = compare(p.SurveyResponses.LifeInsuranceAmount, c)
		case tags.IncomeID:
			c = compare(p.SurveyResponses.HouseholdIncomeAmount, c)
		default:
		}
		comparisons[i] = c
	}
	return comparisons
}

const descrFmt = "You are in the %s percentile of the Hive"
const quartileFmt = "You are in the %s quartile of the Hive"

func compare(val int, c tags.TagComparison) tags.TagComparison {
	var highestPercentile, highestPosition int
	c.Percentiles, highestPercentile, highestPosition = fillPercentiles(val, c.Percentiles)
	c.Percentile = highestPercentile
	c.Value = val
	c.DisplayValue = impart.FriendlyFormatDollars(val)
	c.DisplayPercentile = humanize.Ordinal(c.Percentile)

	//c.DisplayDescription = fmt.Sprintf(descrFmt, strings.TrimSpace(humanize.Ordinal(highestPercentile)))
	c.DisplayDescription = fmt.Sprintf(quartileFmt, strings.TrimSpace(humanize.Ordinal(highestPosition)))
	return c
}

func fillPercentiles(val int, ps tags.Percentiles) (tags.Percentiles, int, int) {
	highestPercentile := 0
	highestPosition := 0
	priorHighValue := 0
	ps.SortAscending()

	for i, p := range ps {
		if val >= p.HighValue || val > priorHighValue {
			p.Fill = true
			highestPercentile = p.Percent
			highestPosition = i + 1
		} else {
			p.Fill = false
		}

		p.DisplayValue = impart.FriendlyFormatDollars(p.HighValue)
		priorHighValue = p.HighValue
		ps[i] = p
	}

	return ps, highestPercentile, highestPosition
}
