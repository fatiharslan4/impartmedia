package profile_test

//
//import (
//	"testing"
//
//	"github.com/gobuffalo/packr"
//	"github.com/impartwealthapp/backend/pkg/impart"
//	"github.com/impartwealthapp/backend/pkg/models"
//	"github.com/impartwealthapp/backend/pkg/profile"
//	"github.com/leebenson/conform"
//	"github.com/segmentio/ksuid"
//	"github.com/stretchr/testify/assert"
//	"github.com/xeipuuv/gojsonschema"
//	"go.uber.org/zap"
//)
//
//var logger, _ = zap.NewDevelopment()
//
//var box = packr.NewBox("./schemas/json")
//var schema, _ = box.FindString("Profile.json")
//var validator = gojsonschema.NewStringLoader(schema)
//
//// moq -pkg profile_test -out profile_dynamo_mock_test.go $GOPATH/src/github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface DynamoDBAPI
//// go get -u github.com/matryer/moq
//// cd $GOPATH/src/github.com/matryer/moq
//// go install
//// moq -pkg profile_test -out internal/pkg/profile/profile_data_mock_test.go pkg/profiledata Store
//func TestNew(t *testing.T) {
//	mockedDynamo := &StoreMock{}
//	p := profile.New(logger.Sugar(), mockedDynamo, impart.NewNoopNotificationService(), validator, "local")
//
//	assert.NotNil(t, p, "expected not nil profile service")
//}
//
////func TestInvalidProfileJson(t *testing.T) {
////	mockedDynamo := &StoreMock{}
////	p := profile.New(logger.Sugar(), mockedDynamo, validator)
////
////
////	_, err := p.NewProfile("abc123", profile.GetProfileInput{ImpartWealthID:"1A"})
////
////	assert.NotNil(t, err)
////	assert.Equal(t, impart.ErrBadRequest, err.Err())
////}
//
////func TestProfileService_GetProfile_New(t *testing.T) {
////	p := models.RandomProfile()
////	profileService := profile.New(logger.Sugar(), &StoreMock{}, validator)
////
////	p, err := profileService.GetProfile(events.APIGatewayProxyRequest{
////		Resource: "/profiles/new",
////	})
////
////	assert.Nil(t, err)
////	ksuid.Parse(p.ImpartWealthID)
////}
//
//func TestProfileService_GetProfile(t *testing.T) {
//	p := models.RandomProfile()
//	profileService := profile.New(logger.Sugar(), &StoreMock{
//		GetProfileFunc: func(impartWealthId string, consistentRead bool) (models.Profile, error) {
//			return p, nil
//		},
//	}, impart.NewNoopNotificationService(), validator, "local")
//
//	testProfile, err := profileService.GetProfile(profile.GetProfileInput{
//		ImpartWealthID: p.ImpartWealthID, ContextAuthID: p.AuthenticationID,
//	})
//
//	assert.NotNil(t, testProfile)
//	assert.Equal(t, p, testProfile)
//	assert.NotEmpty(t, testProfile.Attributes.Name)
//	assert.Nil(t, err)
//}
//
//func TestProfileService_CreateProfile(t *testing.T) {
//	p := models.RandomProfile()
//	profileService := profile.New(logger.Sugar(), &StoreMock{
//		GetProfileFunc: func(impartWealthId string, consistentRead bool) (models.Profile, error) {
//			if impartWealthId == p.ImpartWealthID {
//				return p, nil
//			}
//			return models.Profile{}, impart.ErrNotFound
//		},
//		GetImpartIdFromAuthIdFunc: func(authenticationId string) (string, error) {
//			if authenticationId == p.AuthenticationID {
//				return p.ImpartWealthID, nil
//			}
//			return "", impart.ErrNotFound
//		},
//		GetImpartIdFromEmailFunc: func(email string) (string, error) {
//			if email == p.Email {
//				return p.ImpartWealthID, nil
//			}
//			return "", impart.ErrNotFound
//		},
//		GetImpartIdFromScreenNameFunc: func(screenName string) (string, error) {
//			if screenName == p.ScreenName {
//				return p.ImpartWealthID, nil
//			}
//			return "", impart.ErrNotFound
//		},
//	}, impart.NewNoopNotificationService(), validator, "local")
//
//	testProfile, err := profileService.GetProfile(profile.GetProfileInput{
//		SearchAuthenticationID: p.AuthenticationID, ContextAuthID: p.AuthenticationID,
//	})
//
//	assert.NotNil(t, testProfile)
//	assert.NotEmpty(t, testProfile.Attributes.Name)
//	assert.Nil(t, err)
//
//	testProfile, err = profileService.GetProfile(profile.GetProfileInput{
//		SearchEmail: p.Email, ContextAuthID: p.AuthenticationID,
//	})
//
//	assert.NotNil(t, testProfile)
//	assert.NotEmpty(t, testProfile.Attributes.Name)
//	assert.Nil(t, err)
//
//	testProfile, err = profileService.GetProfile(profile.GetProfileInput{
//		SearchScreenName: p.ScreenName, ContextAuthID: p.AuthenticationID,
//	})
//
//	assert.NotNil(t, testProfile)
//	assert.NotEmpty(t, testProfile.Attributes.Name)
//	assert.Nil(t, err)
//
//}
//
//func TestProfileService_NewProfile(t *testing.T) {
//	p := models.RandomProfile()
//	conform.Strings(&p)
//
//	profileService := profile.New(logger.Sugar(), &StoreMock{
//		CreateProfileFunc: func(p models.Profile) (models.Profile, error) {
//			return p, nil
//		},
//		GetProfileFunc: func(impartWealthId string, consistentRead bool) (models.Profile, error) {
//			if impartWealthId == p.ImpartWealthID {
//				return models.Profile{}, nil
//			}
//			return models.Profile{}, impart.ErrNotFound
//		},
//		GetImpartIdFromAuthIdFunc: func(authenticationId string) (string, error) {
//			if authenticationId == p.AuthenticationID {
//				return "", nil
//			}
//			return "", impart.ErrNotFound
//		},
//		GetImpartIdFromEmailFunc: func(email string) (string, error) {
//			if email == p.Email {
//				return "", nil
//			}
//			return "", impart.ErrNotFound
//		},
//		GetImpartIdFromScreenNameFunc: func(screenName string) (string, error) {
//			if screenName == p.ScreenName {
//				return "", nil
//			}
//			return "", impart.ErrNotFound
//		},
//		GetWhitelistEntryFunc: func(impartWealthID string) (listProfile models.WhiteListProfile, e error) {
//			return models.WhiteListProfile{}, nil
//		},
//		GetHiveFunc: func(hiveID string, consistentRead bool) (hive models.Hive, e error) {
//			hive.HiveID = ksuid.New().String()
//			return
//		},
//	}, impart.NewNoopNotificationService(), validator, "local")
//
//	c, err := profileService.NewProfile(p.AuthenticationID, p)
//
//	assert.Nil(t, err)
//	assert.NotEqual(t, p.CreatedDate, c.CreatedDate)
//	assert.NotEqual(t, p.UpdatedDate, c.UpdatedDate)
//	assert.NotEqual(t, p.Attributes.UpdatedDate, c.Attributes.UpdatedDate)
//	assert.NotEqual(t, p.SurveyResponses.ImportTimestamp, c.SurveyResponses.ImportTimestamp)
//	//assert.True(t, p.EqualsIgnoreTimes(c))
//}
//
//func TestProfileService_NewProfile_Exists(t *testing.T) {
//	p := models.RandomProfile()
//	//req := events.APIGatewayProxyRequest{
//	//	Body: p.ToJson(),
//	//	RequestContext: events.APIGatewayProxyRequestContext{
//	//		Authorizer:  map[string]interface{}{
//	//			"authenticationId": p.AuthenticationID,
//	//		},
//	//	},
//	//}
//
//	profileService := profile.New(logger.Sugar(), &StoreMock{
//		CreateProfileFunc: func(p models.Profile) (models.Profile, error) {
//			return p, nil
//		},
//		GetProfileFunc: func(impartWealthId string, consistentRead bool) (models.Profile, error) {
//			if impartWealthId == p.ImpartWealthID {
//				return p, nil
//			}
//			return models.Profile{}, impart.ErrNotFound
//		},
//		GetWhitelistEntryFunc: func(impartWealthID string) (listProfile models.WhiteListProfile, e error) {
//			return models.WhiteListProfile{}, nil
//		},
//		GetHiveFunc: func(hiveID string, consistentRead bool) (hive models.Hive, e error) {
//			hive.HiveID = ksuid.New().String()
//			return
//		},
//	}, impart.NewNoopNotificationService(), validator, "local")
//
//	_, err := profileService.NewProfile(p.AuthenticationID, p)
//
//	assert.NotNil(t, err)
//	assert.Equal(t, impart.ErrExists, err.Err())
//	assert.Contains(t, err.Msg(), "profile")
//}
//
//func TestProfileService_NewProfile_AuthExists(t *testing.T) {
//	p := models.RandomProfile()
//	//req := events.APIGatewayProxyRequest{
//	//	Body: p.ToJson(),
//	//	RequestContext: events.APIGatewayProxyRequestContext{
//	//		Authorizer:  map[string]interface{}{
//	//			"authenticationId": p.AuthenticationID,
//	//		},
//	//	},
//	//}
//
//	profileService := profile.New(logger.Sugar(), &StoreMock{
//		CreateProfileFunc: func(p models.Profile) (models.Profile, error) {
//			return p, nil
//		},
//		GetProfileFunc: func(impartWealthId string, consistentRead bool) (models.Profile, error) {
//			if impartWealthId == p.ImpartWealthID {
//				return models.Profile{}, nil
//			}
//			return models.Profile{}, impart.ErrNotFound
//		},
//		GetImpartIdFromAuthIdFunc: func(authenticationId string) (string, error) {
//			return p.ImpartWealthID, impart.ErrNotFound
//		},
//		GetWhitelistEntryFunc: func(impartWealthID string) (listProfile models.WhiteListProfile, e error) {
//			return models.WhiteListProfile{}, nil
//		},
//	}, impart.NewNoopNotificationService(), validator, "local")
//
//	_, err := profileService.NewProfile(p.AuthenticationID, p)
//
//	assert.NotNil(t, err)
//	assert.Equal(t, impart.ErrExists, err.Err())
//	assert.Contains(t, err.Msg(), "authenticationId")
//}
//
//func TestProfileService_NewProfile_EmailExists(t *testing.T) {
//	p := models.RandomProfile()
//	//req := events.APIGatewayProxyRequest{
//	//	Body: p.ToJson(),
//	//	RequestContext: events.APIGatewayProxyRequestContext{
//	//		Authorizer:  map[string]interface{}{
//	//			"authenticationId": p.AuthenticationID,
//	//		},
//	//	},
//	//}
//
//	profileService := profile.New(logger.Sugar(), &StoreMock{
//		CreateProfileFunc: func(p models.Profile) (models.Profile, error) {
//			return p, nil
//		},
//		GetProfileFunc: func(impartWealthId string, consistentRead bool) (models.Profile, error) {
//			if impartWealthId == p.ImpartWealthID {
//				return models.Profile{}, nil
//			}
//			return models.Profile{}, impart.ErrNotFound
//		},
//		GetImpartIdFromAuthIdFunc: func(authenticationId string) (string, error) {
//			if authenticationId == p.AuthenticationID {
//				return "", nil
//			}
//			return "", impart.ErrNotFound
//		},
//		GetImpartIdFromEmailFunc: func(email string) (string, error) {
//			return p.ImpartWealthID, nil
//		},
//		GetWhitelistEntryFunc: func(impartWealthID string) (listProfile models.WhiteListProfile, e error) {
//			return models.WhiteListProfile{}, nil
//		},
//	}, impart.NewNoopNotificationService(), validator, "local")
//
//	_, err := profileService.NewProfile(p.AuthenticationID, p)
//
//	assert.NotNil(t, err)
//	assert.Equal(t, impart.ErrExists, err.Err())
//	assert.Contains(t, err.Msg(), "email")
//}
//
//func TestProfileService_NewProfile_ScreenNameExists(t *testing.T) {
//	p := models.RandomProfile()
//	//fmt.Println(p.ToJson())
//	//req := events.APIGatewayProxyRequest{
//	//	Body: p.ToJson(),
//	//	RequestContext: events.APIGatewayProxyRequestContext{
//	//		Authorizer:  map[string]interface{}{
//	//			"authenticationId": p.AuthenticationID,
//	//		},
//	//	},
//	//}
//
//	profileService := profile.New(logger.Sugar(), &StoreMock{
//		CreateProfileFunc: func(p models.Profile) (models.Profile, error) {
//			return p, nil
//		},
//		GetProfileFunc: func(impartWealthId string, consistentRead bool) (models.Profile, error) {
//			if impartWealthId == p.ImpartWealthID {
//				return models.Profile{}, nil
//			}
//			return models.Profile{}, impart.ErrNotFound
//		},
//		GetImpartIdFromAuthIdFunc: func(authenticationId string) (string, error) {
//			if authenticationId == p.AuthenticationID {
//				return "", nil
//			}
//			return "", impart.ErrNotFound
//		},
//		GetImpartIdFromEmailFunc: func(email string) (string, error) {
//			if email == p.Email {
//				return "", nil
//			}
//			return "", impart.ErrNotFound
//		},
//		GetImpartIdFromScreenNameFunc: func(screenName string) (string, error) {
//			return p.ImpartWealthID, nil
//		},
//		GetWhitelistEntryFunc: func(impartWealthID string) (listProfile models.WhiteListProfile, e error) {
//			return models.WhiteListProfile{}, nil
//		},
//	}, impart.NewNoopNotificationService(), validator, "local")
//
//	_, err := profileService.NewProfile(p.AuthenticationID, p)
//
//	assert.NotNil(t, err)
//	assert.Equal(t, impart.ErrExists, err.Err())
//	assert.Contains(t, err.Msg(), "screenName")
//}
