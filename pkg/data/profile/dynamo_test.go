package profile_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	profiledata "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// These tests require a running local dynamoDB environment
// This has been setup in the impart-models repo; you can switch to the "dynamo_schema" folder and execute
// github.com/ImpartWealthApp/impart-models ./run_local.sh
var localDynamo = "http://localhost:8000"
var logger = newLogger()

func newLogger() *zap.SugaredLogger {
	l, _ := zap.NewDevelopment()
	return l.Sugar()
}

type ProfileDynamoSuite struct {
	suite.Suite
	svc          *dynamodb.DynamoDB
	profileTbl   string
	whitelistTbl string
}

func TestProfileDynamoSuite(t *testing.T) {
	suite.Run(t, new(ProfileDynamoSuite))
}

func (s *ProfileDynamoSuite) SetupSuite() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
	})
	if err != nil {
		s.FailNow(err.Error())
	}
	s.svc = dynamodb.New(sess, &aws.Config{Endpoint: aws.String(localDynamo)})
	s.profileTbl = fmt.Sprintf("%s_%s", "local", "profile")
	s.whitelistTbl = fmt.Sprintf("%s_%s", "local", "whitelist_profile")
}

func (s *ProfileDynamoSuite) TruncateLocalTables() {
	for _, tbl := range []string{s.whitelistTbl, s.profileTbl} {
		params := &dynamodb.ScanInput{
			TableName: &tbl,
		}
		result, err := s.svc.Scan(params)
		if err != nil {
			s.FailNow(err.Error())
		}
		if result == nil || result.Items == nil {
			s.FailNow("result or item map is nil")
		}
		for _, i := range result.Items {
			input := &dynamodb.DeleteItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"impartWealthId": {
						S: i["impartWealthId"].S,
					},
				},
				TableName: &tbl,
			}

			_, err := s.svc.DeleteItem(input)
			if err != nil {
				s.FailNow(err.Error())
			}
		}
	}
}

func (s *ProfileDynamoSuite) SetupTest() {
	s.TruncateLocalTables()
}

func (s *ProfileDynamoSuite) TestProfileDynamoDb_CreateProfile() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)

	randomProfile := models.RandomProfile()
	randomProfile.Attributes.HiveMemberships = models.HiveMemberships{
		models.HiveMembership{
			HiveID: ksuid.New().String(),
		},
	}

	updatedProfile, err := pd.UpdateProfile(randomProfile.AuthenticationID, randomProfile)
	s.Empty(updatedProfile)
	s.NotNil(err)
	s.Equal(impart.ErrNotFound, err)

	createdProfile, err := pd.CreateProfile(randomProfile)
	s.NoError(err)

	s.Equal(randomProfile.UpdatedDate, createdProfile.UpdatedDate)
	s.Equal(randomProfile, createdProfile)

}

func (s *ProfileDynamoSuite) TestProfileDynamoDb_UpdateProfile() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)

	randomProfile := models.RandomProfile()

	createdProfile, err := pd.CreateProfile(randomProfile)
	s.NoError(err)
	s.Equal(randomProfile, createdProfile)

	createdProfile.Email = createdProfile.Email + "1"

	updatedProfile, err := pd.UpdateProfile(randomProfile.AuthenticationID, createdProfile)
	s.NoError(err)
	s.Equal(createdProfile, updatedProfile)

}

func (s *ProfileDynamoSuite) TestProfileDynamoDb_UpdateProfileDoesNotExist() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)

	randomProfile := models.RandomProfile()

	createdProfile, err := pd.CreateProfile(randomProfile)
	s.NoError(err)
	s.Equal(randomProfile, createdProfile)

	createdProfile.ImpartWealthID = "idon'texist"

	updatedProfile, err := pd.UpdateProfile(randomProfile.AuthenticationID, createdProfile)
	s.Error(err)
	s.Equal(impart.ErrNotFound, err)
	s.Empty(updatedProfile)

}

func (s *ProfileDynamoSuite) TestProfileDynamoDb_UpdateProfileBadAuthId() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)

	randomProfile := models.RandomProfile()

	createdProfile, err := pd.CreateProfile(randomProfile)
	s.NoError(err)
	s.Equal(randomProfile, createdProfile)

	createdProfile.ImpartWealthID = "idon'texist"

	updatedProfile, err := pd.UpdateProfile("badauthid", createdProfile)
	s.Error(err)
	s.Equal(impart.ErrNotFound, err)
	s.Empty(updatedProfile)

}

func (s *ProfileDynamoSuite) TestLookups() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)

	randomProfile := models.RandomProfile()

	createdProfile, err := pd.CreateProfile(randomProfile)
	s.NoError(err)
	s.Equal(randomProfile, createdProfile)

	authIdImpartId, err := pd.GetImpartIdFromAuthId(createdProfile.AuthenticationID)
	s.NoError(err)
	s.Equal(createdProfile.ImpartWealthID, authIdImpartId)

	emailImpartId, err := pd.GetImpartIdFromEmail(createdProfile.Email)
	s.NoError(err)
	s.Equal(createdProfile.ImpartWealthID, emailImpartId)

	screenNameImpartId, err := pd.GetImpartIdFromScreenName(createdProfile.ScreenName)
	s.NoError(err)
	s.Equal(createdProfile.ImpartWealthID, screenNameImpartId)
}

func (s *ProfileDynamoSuite) TestMissing() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)

	randomProfile := models.RandomProfile()

	p, err := pd.GetProfile(randomProfile.ImpartWealthID, false)
	s.NotNil(err)
	s.Equal(impart.ErrNotFound, err)
	s.Empty(p)

	impartId, err := pd.GetImpartIdFromAuthId(randomProfile.AuthenticationID)
	s.NotNil(err)
	s.Equal(impart.ErrNotFound, err)
	s.Empty(impartId)

}

func (s *ProfileDynamoSuite) TestDoesNotExist() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)

	randomProfile := models.RandomProfile()

	impartId, err := pd.GetProfile(randomProfile.ImpartWealthID, false)
	s.NotNil(err)
	s.Equal(impart.ErrNotFound, err)
	s.Empty(impartId)
}

func (s *ProfileDynamoSuite) TestProfileDynamoDb_GetProfileFromAuthId() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)

	randomProfile := models.RandomProfile()

	createdProfile, err := pd.CreateProfile(randomProfile)
	s.NoError(err)
	s.Equal(randomProfile, createdProfile)

	authProfile, err := pd.GetProfileFromAuthId(randomProfile.AuthenticationID, true)
	s.NoError(err)
	s.Equal(randomProfile, authProfile)

}

func (s *ProfileDynamoSuite) TestProfileDynamoDb_DeleteProfile() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)

	randomProfile := models.RandomProfile()

	updatedProfile, err := pd.UpdateProfile(randomProfile.AuthenticationID, randomProfile)
	s.Empty(updatedProfile)
	s.NotNil(err)
	s.Equal(impart.ErrNotFound, err)

	createdProfile, err := pd.CreateProfile(randomProfile)
	s.NoError(err)
	s.Equal(randomProfile, createdProfile)

	err = pd.DeleteProfile(createdProfile.ImpartWealthID)
	s.NoError(err)

	_, err = pd.GetProfile(createdProfile.ImpartWealthID, true)
	s.Error(err)
	s.Equal(impart.ErrNotFound, err)

}

func (s *ProfileDynamoSuite) TestProfileDynamoDb_UpdateProfileProperty() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)

	randomProfile := models.RandomProfile()

	var empty = models.NotificationProfile{}
	createdProfile, err := pd.CreateProfile(randomProfile)
	s.NoError(err)
	s.Equal(empty, createdProfile.NotificationProfile)

	np := models.NotificationProfile{
		DeviceToken: ksuid.New().String(),
	}

	err = pd.UpdateProfileProperty(randomProfile.ImpartWealthID, "notificationProfile", np)
	s.NoError(err)

	updatedProfile, err := pd.GetProfile(randomProfile.ImpartWealthID, true)
	s.NoError(err)
	s.Equal(np.DeviceToken, updatedProfile.NotificationProfile.DeviceToken)
}

func (s *ProfileDynamoSuite) TestProfileDynamoDb_GetNotificationProfiles() {
	deleteProfiles()
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)

	p1 := models.RandomProfile()
	var empty = models.NotificationProfile{}
	p1.NotificationProfile = empty

	p2 := models.RandomProfile()
	p2.NotificationProfile.DeviceToken = "sometoken"

	p3 := models.RandomProfile()
	p3.NotificationProfile.AWSPlatformEndpointARN = "somearn"

	cp1, err := pd.CreateProfile(p1)
	s.NoError(err)
	s.Equal(empty, cp1.NotificationProfile)
	cp2, err := pd.CreateProfile(p2)
	s.NoError(err)
	s.Equal("sometoken", cp2.NotificationProfile.DeviceToken)
	cp3, err := pd.CreateProfile(p3)
	s.NoError(err)
	s.Equal("somearn", cp3.NotificationProfile.AWSPlatformEndpointARN)

	// Assert
	profiles, nextPage, err := pd.GetNotificationProfiles(nil)
	s.NoError(err)
	s.Len(profiles, 1)
	s.Nil(nextPage)
	s.Equal("sometoken", profiles[0].NotificationProfile.DeviceToken)

}

func deleteProfiles() {
	pd, _ := profiledata.New("us-east-2", localDynamo, "local", logger)
	profiles, nextPage, _ := pd.GetNotificationProfiles(nil)

	for _, p := range profiles {
		pd.DeleteProfile(p.ImpartWealthID)
	}

	for nextPage != nil {
		for _, p := range profiles {
			pd.DeleteProfile(p.ImpartWealthID)
		}
	}

}
