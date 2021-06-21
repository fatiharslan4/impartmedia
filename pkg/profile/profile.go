package profile

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/impartwealthapp/backend/pkg/models/dbmodels"

	"github.com/volatiletech/sqlboiler/v4/boil"

	profile_data "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/data/types"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
)

type Service interface {
	QuestionnaireService
	NewProfile(ctx context.Context, p models.Profile) (models.Profile, impart.Error)
	GetProfile(ctx context.Context, getProfileInput GetProfileInput) (models.Profile, impart.Error)
	UpdateProfile(ctx context.Context, p models.Profile) (models.Profile, impart.Error)
	DeleteProfile(ctx context.Context, impartWealthID string) impart.Error
	ScreenNameExists(ctx context.Context, screenName string) bool

	ValidateSchema(document gojsonschema.JSONLoader) []impart.Error
	ValidateScreenNameInput(document gojsonschema.JSONLoader) []impart.Error
	ValidateScreenNameString(ctx context.Context, screenName string) impart.Error
	ValidateInput(document gojsonschema.JSONLoader, validationModel types.Type) []impart.Error
	Logger() *zap.Logger

	ModifyUserConfigurations(ctx context.Context, conf models.UserConfigurations) (models.UserConfigurations, impart.Error)
	GetUserConfigurations(ctx context.Context, impartWealthID string) (models.UserConfigurations, impart.Error)

	GetUserDevice(ctx context.Context, token string, impartWealthID string, deviceToken string) (models.UserDevice, error)
	CreateUserDevice(ctx context.Context, user *dbmodels.User, ud *dbmodels.UserDevice) (models.UserDevice, impart.Error)

	MapDeviceForNotification(ctx context.Context, ud models.UserDevice) impart.Error
	UpdateExistingNotificationMappData(input models.MapArgumentInput, notifyStatus bool) impart.Error

	BlockUser(ctx context.Context, impartWealthID string, screenname string, status bool) impart.Error
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
	var deviceToken string

	contextAuthId := impart.GetCtxAuthID(ctx)
	if contextAuthId == "" {
		return empty, impart.NewError(impart.ErrBadRequest, "Unable to locate authenticationID")
	}

	// check device token provided
	//  check the device token is provided with either
	// input deviceToken / with userDevices
	deviceToken = p.DeviceToken
	if (len(p.UserDevices) > 0 && p.UserDevices[0] != models.UserDevice{}) {
		deviceToken = p.UserDevices[0].DeviceID
	}

	//
	// If device token is not found from input
	//
	if deviceToken == "" {
		ps.Logger().Debug("Unable to locate device id",
			zap.Any("profile", p),
		)
	}

	ctxUser, err := ps.profileStore.GetUserFromAuthId(ctx, contextAuthId)
	if err != nil {
		if err == impart.ErrNotFound {
			//new authenticated user is created this profile/user
			p.AuthenticationID = contextAuthId
			ps.Logger().Debug("requested to create new profile",
				zap.Any("AuthenticationID", p.AuthenticationID))
		} else {
			ps.Logger().Error("error checking existing profile", zap.Error(err))
			return empty, impart.NewError(impart.ErrUnknown, "unable to check existing profile")
		}
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
		ps.Logger().Error("create of profile failed - unexpected situation", zap.Any("contextUser", ctxUser), zap.Any("inputProfile", p))
		return empty, impart.NewError(impart.ErrUnknown, "unable to create profile; unknown state")
	}

	if impartErr := ps.validateNewProfile(ctx, p); impartErr != nil {
		return empty, impartErr
	}

	ps.Logger().Debug("creating a new user profile", zap.Any("updated", p))
	p.CreatedDate = impart.CurrentUTC()
	p.UpdatedDate = impart.CurrentUTC()
	p.Attributes.UpdatedDate = impart.CurrentUTC()
	p.Admin = false

	dbUser, err := p.DBUser()
	if err != nil {
		return empty, impart.NewError(impart.ErrUnknown, "couldn't convert profile to profileStore user")
	}
	dbProfile, err := p.DBProfile()
	if err != nil {
		return empty, impart.NewError(impart.ErrUnknown, "couldn't convert profile to profileStore profile")
	}
	dbUser.CreatedAt = impart.CurrentUTC()
	dbUser.UpdatedAt = impart.CurrentUTC()
	dbProfile.UpdatedAt = impart.CurrentUTC()

	// hide this when new notify workflow ok : begin
	endpointARN, err := ps.notificationService.SyncTokenEndpoint(ctx, p.DeviceToken, "")
	if err != nil {
		ps.Logger().Error("Token Sync Endpoint error", zap.Any("Error", err), zap.Any("contextUser", ctxUser), zap.Any("inputProfile", p))
	}
	dbUser.AwsSNSAppArn = endpointARN
	// hide this : end

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

	// save notification configuration
	// based on user added notification status
	// insert entry to table
	var notificationStatus bool
	if (p.Settings != models.UserSettings{} && p.Settings.NotificationStatus) {
		notificationStatus = true
	}
	_, err = ps.ModifyUserConfigurations(ctx, models.UserConfigurations{
		ImpartWealthID:     dbUser.ImpartWealthID,
		NotificationStatus: notificationStatus,
	})
	if err != nil {
		ps.Logger().Error("unable to process your request", zap.Error(err))
	}
	out.Settings.NotificationStatus = notificationStatus

	//
	// Register the device for notification
	//
	if (len(p.UserDevices) > 0 && p.UserDevices[0] != models.UserDevice{}) {
		// check empty device id but provided with devide token
		if p.UserDevices[0].DeviceToken == "" && p.DeviceToken != "" {
			p.UserDevices[0].DeviceToken = p.DeviceToken
		}
		// check the device id exists
		if p.UserDevices[0].DeviceToken != "" {
			userDevice, err := ps.CreateUserDevice(ctx, dbUser, p.UserDevices[0].UserDeviceToDBModel())
			if err != nil {
				impartErr := impart.NewError(impart.ErrBadRequest, fmt.Sprintf("unable to add/update the device information %v", err))
				ps.Logger().Error(impartErr.Error())
			}

			// map for notification
			err = ps.MapDeviceForNotification(ctx, userDevice)
			if err != nil {
				impartErr := impart.NewError(impart.ErrBadRequest, fmt.Sprintf("an error occured in update mapping for notification %v", err))
				ps.Logger().Error(impartErr.Error())
			}

			out.UserDevices = append(out.UserDevices, userDevice)
		} else {
			impartErr := impart.NewError(impart.ErrBadRequest, fmt.Sprintf("can't register device, device id not found %v", err))
			ps.Logger().Error(impartErr.Error(), zap.Any("data", p))
		}
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
		if err != impart.ErrNotFound {
			return models.Profile{}, impart.NewError(err, "error fetching matching profile")
		}
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
		p.UpdatedDate.IsZero() || !p.CreatedDate.Equal(existingDBUser.CreatedAt) || p.UpdatedDate.Sub(existingDBUser.UpdatedAt) < 0 {
		msg := "profile being updated appears to be incorrect - critical properties do not match the profile being updated."
		ps.Logger().Error(msg, zap.Time("inCreatedAt", p.CreatedDate),
			zap.Time("existingCreatedAt", existingDBUser.CreatedAt),
			zap.Time("inUpdatedAt", p.UpdatedDate),
			zap.Time("existingUpdatedAt", existingDBUser.UpdatedAt))
		return empty, impart.NewError(impart.ErrBadRequest, msg)
	}

	if existingDBUser.DeviceToken != "" && strings.TrimSpace(p.DeviceToken) == "" {
		err = ps.notificationService.UnsubscribeAll(ctx, existingDBUser.ImpartWealthID)
		ps.Logger().Error("error unsusbcribing", zap.Error(err))
	} else if existingDBUser.DeviceToken != p.DeviceToken {
		existingDBUser.DeviceToken = p.DeviceToken
		existingDBUser.UpdatedAt = impart.CurrentUTC()
		if err := ps.SubscribeNewDeviceToken(ctx, existingDBUser); err != nil {
			return empty, impart.NewError(impart.ErrUnknown, "unknown error updating subscriptions")
		}
	}

	tmpProfile, err := models.ProfileFromDBModel(existingDBUser, existingDBProfile)
	if err != nil {
		return models.Profile{}, impart.NewError(impart.ErrUnknown, "unable to generate an impart profile from existing DB profile")
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
	user.UpdatedAt = impart.CurrentUTC()
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
