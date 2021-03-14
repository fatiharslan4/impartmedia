package profile

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"reflect"
	"strings"

	profile_data "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
)

type Service interface {
	NewProfile(ctx context.Context, p models.Profile) (models.Profile, impart.Error)
	GetProfile(ctx context.Context, getProfileInput GetProfileInput) (models.Profile, impart.Error)
	UpdateProfile(ctx context.Context, p models.Profile) (models.Profile, impart.Error)
	DeleteProfile(ctx context.Context, impartWealthID string) impart.Error
	ScreenNameExists(ctx context.Context, screenName string) bool

	ValidateSchema(document gojsonschema.JSONLoader) impart.Error
	Logger() *zap.Logger
}

func New(logger *zap.SugaredLogger, db *sql.DB, dal profile_data.Store, ns impart.NotificationService, schema gojsonschema.JSONLoader, stage string) Service {
	return &profileService{
		stage:               stage,
		SugaredLogger:       logger,
		profileStore:        dal,
		notificationService: ns,
		schemaValidator:     schema,
		db:                  db,
	}
}

type profileService struct {
	stage string
	*zap.SugaredLogger
	profileStore        profile_data.Store
	schemaValidator     gojsonschema.JSONLoader
	notificationService impart.NotificationService
	db                  *sql.DB
}

func (ps *profileService) ScreenNameExists(ctx context.Context, screenName string) bool {
	user, err := ps.profileStore.GetUserFromScreenName(ctx, screenName)
	if err == nil && user != nil {
		return true
	}
	return false
}

func (ps *profileService) Logger() *zap.Logger {
	return ps.Desugar()
}

func (ps *profileService) DeleteProfile(ctx context.Context, impartWealthID string) impart.Error {
	if strings.TrimSpace(impartWealthID) == "" {
		return impart.NewError(impart.ErrBadRequest, "impartWealthID is empty")
	}
	contextUser := impart.GetCtxUser(ctx)
	if contextUser == nil || contextUser.ImpartWealthID == "" {
		return impart.NewError(impart.ErrBadRequest, "context user not found")
	}
	userToDelete, err := ps.profileStore.GetUser(ctx, impartWealthID)
	if err != nil {
		return impart.NewError(err, fmt.Sprintf("couldn't find profile for impartWealthID %s", impartWealthID))
	}

	if contextUser.ImpartWealthID == userToDelete.ImpartWealthID ||
		contextUser.Admin {
		ps.Logger().Info("request to delete a user passed validation", zap.String("deleteUser", userToDelete.ImpartWealthID),
			zap.String("contextUser", contextUser.ImpartWealthID))
		err = ps.profileStore.DeleteProfile(ctx, impartWealthID)
		if err != nil {
			return impart.NewError(err, "unable to retrieve profile")
		}
		return nil
	}

	return impart.NewError(impart.ErrUnauthorized, "user is not authorized")

}

func (ps *profileService) NewProfile(ctx context.Context, p models.Profile) (models.Profile, impart.Error) {
	var empty models.Profile
	var err error

	contextAuthId := impart.GetCtxAuthID(ctx)
	if contextAuthId == "" {
		return empty, impart.NewError(impart.ErrBadRequest, "Unable to locate authenticationID")
	}
	ctxUser, err := ps.profileStore.GetUserFromAuthId(ctx, contextAuthId)
	if err != nil && err != impart.ErrNotFound {
		ps.Logger().Error("error checking existing profile", zap.Error(err))
		return empty, impart.NewError(impart.ErrUnknown, "unable to check existing profile")
	} else if err == impart.ErrNotFound {
		//new authenticated user is created this profile/user
		p.AuthenticationID = contextAuthId
	} else if ctxUser != nil && ctxUser.Admin {
		//allow the creation of the profile by an admin
		ps.Logger().Info("admin user is creating a new user",
			zap.String("admin", ctxUser.Email),
			zap.String("userEmail", p.ImpartWealthID),
			zap.String("userEmail", p.Email))
	} else if ctxUser.ImpartWealthID == p.ImpartWealthID {
		return empty, impart.NewError(impart.ErrExists, "an existing impart wealth user for this id exists")
	} else {
		//what?
		return empty, impart.NewError(impart.ErrUnknown, "unable to create profile; unknown state")
	}

	p.SurveyResponses = models.SurveyResponses{}

	if impartErr := ps.validateNewProfile(ctx, p); impartErr != nil {
		return empty, impartErr
	}

	p.CreatedDate = impart.CurrentUTC()
	p.UpdatedDate = impart.CurrentUTC()
	p.Attributes.UpdatedDate = impart.CurrentUTC()
	p.SurveyResponses.ImportTimestamp = impart.CurrentUTC()
	p.Admin = false

	dbUser, err := p.DBUser()
	if err != nil {
		return empty, impart.NewError(impart.ErrUnknown, "couldn't convert profile to profileStore user")
	}
	dbProfile, err := p.DBProfile()
	if err != nil {
		return empty, impart.NewError(impart.ErrUnknown, "couldn't convert profile to profileStore profile")
	}
	dbUser.CreatedTS = impart.CurrentUTC()
	dbUser.UpdatedTS = impart.CurrentUTC()
	dbProfile.UpdatedTS = impart.CurrentUTC()
	endpointARN, err := ps.notificationService.SyncTokenEndpoint(ctx, p.DeviceToken, "")
	dbUser.AwsSNSAppArn = endpointARN

	err = ps.profileStore.CreateUserProfile(ctx, dbUser, dbProfile)
	if err != nil {
		ps.Error(err)
		return empty, impart.NewError(err, "unable to create profile")
	}
	dbUser, err = ps.profileStore.GetUser(ctx, dbUser.ImpartWealthID)
	if err != nil {
		ps.Error(err)
		return empty, impart.NewError(err, "unable to create profile")
	}
	dbProfile = dbUser.R.ImpartWealthProfile
	out, err := models.ProfileFromDBModel(dbUser, dbProfile)
	if err != nil {
		return models.Profile{}, impart.NewError(err, "")
	}

	return *out, nil
}

type GetProfileInput struct {
	ImpartWealthID, SearchEmail, SearchScreenName string
}

func (gpi GetProfileInput) isSelf(ctxUser *dbmodels.User) bool {
	if gpi.ImpartWealthID != "" && gpi.ImpartWealthID == ctxUser.ImpartWealthID {
		return true
	}
	if gpi.SearchEmail != "" && gpi.SearchEmail == ctxUser.Email {
		return true
	}
	if gpi.SearchScreenName != "" && gpi.SearchScreenName == ctxUser.ScreenName {
		return true
	}
	return false
}

func (ps *profileService) GetProfile(ctx context.Context, gpi GetProfileInput) (models.Profile, impart.Error) {
	var out *models.Profile
	var u *dbmodels.User
	var err error

	ctxUser := impart.GetCtxUser(ctx)
	if gpi.isSelf(ctxUser) {
		u = ctxUser
	} else if gpi.ImpartWealthID != "" {
		u, err = ps.profileStore.GetUser(ctx, gpi.ImpartWealthID)
	} else if gpi.SearchEmail != "" {
		u, err = ps.profileStore.GetUserFromEmail(ctx, gpi.SearchEmail)
	} else if gpi.SearchScreenName != "" {
		u, err = ps.profileStore.GetUserFromScreenName(ctx, gpi.SearchScreenName)
	}
	if err != nil {
		return models.Profile{}, impart.NewError(err, "unable to find a matching profile")
	}

	p := u.R.ImpartWealthProfile
	if p == nil {
		p, err = ps.profileStore.GetProfile(ctx, u.ImpartWealthID)
	}
	if err != nil {
		return models.Profile{}, impart.NewError(err, "unable to find a matching profile")
	}

	out, err = models.ProfileFromDBModel(u, p)
	if err != nil {
		return models.Profile{}, impart.NewError(err, "")
	}

	return *out, nil
}

func (ps *profileService) UpdateProfile(ctx context.Context, p models.Profile) (models.Profile, impart.Error) {
	var empty models.Profile
	var err error

	ctxUser := impart.GetCtxUser(ctx)

	if p.AuthenticationID != ctxUser.AuthenticationID && !ctxUser.Admin || p.ImpartWealthID != ctxUser.ImpartWealthID {
		return models.Profile{}, impart.NewError(impart.ErrUnauthorized, fmt.Sprintf("user %s is not authorized to modify profile of %s: %s",
			ctxUser.ScreenName, p.ScreenName, p.ImpartWealthID))
	}

	existingDBUser, err := ps.profileStore.GetUser(ctx, p.ImpartWealthID)
	if err != nil {
		return models.Profile{}, impart.NewError(impart.ErrUnknown, "unable to fetch existing user from dB")
	}
	existingDBProfile := existingDBUser.R.ImpartWealthProfile
	ps.Logger().Debug("Checking Updated Profile",
		zap.Any("existingDBUser", *existingDBUser),
		zap.Any("existingDBProfile", *existingDBProfile),
		zap.Any("updated", p))

	if p.CreatedDate.IsZero() ||
		p.UpdatedDate.IsZero() || !p.CreatedDate.Equal(existingDBUser.CreatedTS) || !p.UpdatedDate.Equal(existingDBUser.UpdatedTS) {
		msg := "profile being updated appears to be incorrect - critical properties do not match the profile being updated."
		ps.Logger().Error(msg, zap.Time("inCreatedTS", p.CreatedDate),
			zap.Time("existingCreatedTS", existingDBUser.CreatedTS),
			zap.Time("inUpdatedTS", p.UpdatedDate),
			zap.Time("existingUpdatedTS", existingDBUser.UpdatedTS))
		return empty, impart.NewError(impart.ErrBadRequest, msg)
	}

	if existingDBUser.DeviceToken != "" && strings.TrimSpace(p.DeviceToken) == "" {
		err = ps.notificationService.UnsubscribeAll(ctx, existingDBUser.ImpartWealthID)
		ps.Logger().Error("error unsusbcribing", zap.Error(err))
	} else if existingDBUser.DeviceToken != p.DeviceToken {
		existingDBUser.DeviceToken = p.DeviceToken
		existingDBUser.UpdatedTS = impart.CurrentUTC()
		if err := ps.SubscribeNewDeviceToken(ctx, existingDBUser); err != nil {
			return empty, impart.NewError(impart.ErrUnknown, "unknown error updating subscriptions")
		}
	}

	tmpProfile, err := models.ProfileFromDBModel(existingDBUser, existingDBProfile)
	if err != nil {
		return models.Profile{}, impart.NewError(impart.ErrUnknown, "unable to generate an impart profile from existing DB profile")
	}

	if !reflect.DeepEqual(p.SurveyResponses, tmpProfile.SurveyResponses) {
		if err := existingDBProfile.SurveyResponses.Marshal(p.SurveyResponses); err != nil {
			return empty, impart.NewError(impart.ErrUnknown, "unable to update profile")
		}
	}

	if !reflect.DeepEqual(p.Attributes, tmpProfile.Attributes) {
		if err := existingDBProfile.Attributes.Marshal(p.Attributes); err != nil {
			return empty, impart.NewError(impart.ErrUnknown, "unable to update profile")
		}
	}

	if p.ScreenName != existingDBUser.ScreenName {
		existingDBUser.ScreenName = p.ScreenName
	}

	err = ps.profileStore.UpdateProfile(ctx, existingDBUser, existingDBProfile)
	if err != nil {
		ps.Logger().Error("couldn't save profile in DB", zap.Error(err))
		return empty, impart.NewError(impart.ErrUnknown, "unable save profile in DB")
	}

	up, err := models.ProfileFromDBModel(existingDBUser, existingDBProfile)
	if err != nil {
		return empty, impart.NewError(impart.ErrUnknown, "unable to update profile")
	}
	return *up, nil
}

func (ps *profileService) SubscribeNewDeviceToken(ctx context.Context, user *dbmodels.User) error {
	endpointARN, err := ps.notificationService.SyncTokenEndpoint(ctx, user.DeviceToken, user.AwsSNSAppArn)
	if err != nil {
		return err
	}
	user.AwsSNSAppArn = endpointARN
	user.UpdatedTS = impart.CurrentUTC()
	if _, err := user.Update(ctx, ps.db, boil.Infer()); err != nil {
		return err
	}

	subs, err := dbmodels.NotificationSubscriptions(
		dbmodels.NotificationSubscriptionWhere.ImpartWealthID.EQ(user.ImpartWealthID)).All(ctx, ps.db)

	for _, sub := range subs {
		if err := ps.notificationService.SubscribeTopic(ctx, user.ImpartWealthID, sub.TopicArn); err != nil {
			return err
		}
	}
	return nil
}
