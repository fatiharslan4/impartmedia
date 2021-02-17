package profile_test

import (
	profiledata "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/models"
)

func (s *ProfileDynamoSuite) TestProfileDynamoDb_CreateWhitelistEntry() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)
	wp := models.WhiteListProfile{}
	wp.Randomize()
	err := pd.CreateWhitelistEntry(wp)
	s.NoError(err)

}

func (s *ProfileDynamoSuite) TestProfileDynamoDb_GetWhitelistEntry() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)
	wp := models.WhiteListProfile{}
	wp.Randomize()
	err := pd.CreateWhitelistEntry(wp)
	s.NoError(err)

	wpOut, err := pd.GetWhitelistEntry(wp.ImpartWealthID)
	s.NoError(err)
	s.Equal(wp, wpOut)
}

func (s *ProfileDynamoSuite) TestProfileDynamoDb_SearchWhitelistEntry() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)
	wp := models.WhiteListProfile{}
	wp.Randomize()
	err := pd.CreateWhitelistEntry(wp)
	s.NoError(err)

	wpOut, err := pd.GetWhitelistEntry(wp.ImpartWealthID)
	s.NoError(err)
	s.Equal(wp, wpOut)

	wpOut, err = pd.SearchWhitelistEntry(profiledata.EmailSearchType(), wp.Email)
	s.NoError(err)
	s.Equal(wp, wpOut)

	wpOut, err = pd.SearchWhitelistEntry(profiledata.ScreenNameSearchType(), wp.ScreenName)
	s.NoError(err)
	s.Equal(wp, wpOut)
}

func (s *ProfileDynamoSuite) TestProfileDynamoDb_UpdateWhitelistEntryScreenName() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)
	wp := models.WhiteListProfile{}
	wp.Randomize()
	err := pd.CreateWhitelistEntry(wp)
	s.NoError(err)

	newScreenName := "abc123"

	err = pd.UpdateWhitelistEntryScreenName(wp.ImpartWealthID, newScreenName)
	s.NoError(err)

	wpOut, err := pd.GetWhitelistEntry(wp.ImpartWealthID)
	s.NoError(err)
	s.Equal(newScreenName, wpOut.ScreenName)
}

func (s *ProfileDynamoSuite) TestProfileDynamoDb_GetEmail() {
	email := "dev.poster@test.com"
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)
	wp := models.WhiteListProfile{}
	wp.Randomize()
	wp.Email = email
	err := pd.CreateWhitelistEntry(wp)
	s.NoError(err)

	wpOut, err := pd.GetWhitelistEntry(wp.ImpartWealthID)
	s.NoError(err)
	s.Equal(email, wpOut.Email)

	emailSearchedProfile, err := pd.SearchWhitelistEntry(profiledata.EmailSearchType(), email)
	s.NoError(err)
	s.Equal(email, emailSearchedProfile.Email)
}
