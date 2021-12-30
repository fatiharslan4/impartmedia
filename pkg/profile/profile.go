package profile

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"

	"github.com/beeker1121/mailchimp-go/lists/members"
	hive_data "github.com/impartwealthapp/backend/pkg/data/hive"
	profile_data "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/data/types"
	hive_main "github.com/impartwealthapp/backend/pkg/hive"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
)

type Service interface {
	QuestionnaireService
	NewProfile(ctx context.Context, p models.Profile, apiVersion string) (models.Profile, impart.Error)
	GetProfile(ctx context.Context, getProfileInput GetProfileInput) (models.Profile, impart.Error)
	UpdateProfile(ctx context.Context, p models.Profile) (models.Profile, impart.Error)
	DeleteProfile(ctx context.Context, impartWealthID string, hardtDelete bool, deleteUser models.DeleteUserInput) impart.Error
	ScreenNameExists(ctx context.Context, screenName string) bool

	ValidateSchema(document gojsonschema.JSONLoader) []impart.Error
	ValidateScreenNameInput(document gojsonschema.JSONLoader, scrnName []byte) []impart.Error
	ValidateScreenNameString(ctx context.Context, screenName string) impart.Error
	ValidateInput(document gojsonschema.JSONLoader, validationModel types.Type) []impart.Error
	Logger() *zap.Logger

	ModifyUserConfigurations(ctx context.Context, conf models.UserConfigurations) (models.UserConfigurations, impart.Error)
	GetUserConfigurations(ctx context.Context, impartWealthID string) (models.UserConfigurations, impart.Error)

	GetUserDevice(ctx context.Context, token string, impartWealthID string, deviceToken string) (models.UserDevice, error)
	CreateUserDevice(ctx context.Context, user *dbmodels.User, ud *dbmodels.UserDevice) (models.UserDevice, impart.Error)
	UpdateDeviceToken(ctx context.Context, token string, deviceToken string) impart.Error
	DeleteExceptUserDevice(ctx context.Context, impartID string, deviceToken string, refToken string) error

	MapDeviceForNotification(ctx context.Context, ud models.UserDevice, isAdmin bool) impart.Error
	UpdateExistingNotificationMappData(input models.MapArgumentInput, notifyStatus bool) impart.Error
	BlockUser(ctx context.Context, impartWealthID string, screenname string, status bool) impart.Error

	GetHive(ctx context.Context, hiveID uint64) (*dbmodels.Hive, impart.Error)
	UpdateReadCommunity(ctx context.Context, p models.UpdateReadCommunity, impartID string) impart.Error
	GetUsersDetails(ctx context.Context, gpi models.GetAdminInputs) ([]models.UserDetail, *models.NextPage, impart.Error)
	GetPostDetails(ctx context.Context, gpi models.GetAdminInputs) ([]models.PostDetail, *models.NextPage, impart.Error)
	EditUserDetails(ctx context.Context, gpi models.WaitListUserInput) (string, impart.Error)
	EditBulkUserDetails(ctx context.Context, gpi models.UserUpdate) (*models.UserUpdate, impart.Error)

	DeleteUserByAdmin(ctx context.Context, hardtDelete bool, deleteUser models.DeleteUserInput) impart.Error
	GetHiveDetails(ctx context.Context, gpi models.GetAdminInputs) (models.HiveDetails, *models.NextPage, impart.Error)
	GetFilterDetails(ctx context.Context) ([]byte, impart.Error)

	CreatePlaidProfile(ctx context.Context, plaid models.PlaidInput) (models.PlaidInput, impart.Error)

	GetWeeklyNotification(ctx context.Context)
	GetWeeklyMostPopularNotification(ctx context.Context)
	UpdateUserDevicesDetails(ctx context.Context, userDevice *dbmodels.UserDevice, login bool) (bool, error)
	GetHiveNotification(ctx context.Context) error

	UserEmailDetailsUpdate(ctx context.Context, gpi models.WebAppUserInput) impart.Error
}

func New(logger *zap.SugaredLogger, db *sql.DB, dal profile_data.Store, ns impart.NotificationService, schema gojsonschema.JSONLoader, stage string, hivedata hive_main.Service, hiveSotre hive_data.Hives) Service {
	return &profileService{
		stage:               stage,
		SugaredLogger:       logger,
		profileStore:        dal,
		notificationService: ns,
		schemaValidator:     schema,
		db:                  db,
		hiveData:            hivedata,
		hiveStore:           hiveSotre,
	}
}

type profileService struct {
	stage string
	*zap.SugaredLogger
	profileStore        profile_data.Store
	schemaValidator     gojsonschema.JSONLoader
	notificationService impart.NotificationService
	db                  *sql.DB
	hiveData            hive_main.Service
	hiveStore           hive_data.Hives
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

func (ps *profileService) DeleteProfile(ctx context.Context, impartWealthID string, hardDelete bool, deleteUser models.DeleteUserInput) impart.Error {
	if strings.TrimSpace(impartWealthID) == "" {
		return impart.NewError(impart.ErrBadRequest, "impartWealthID is empty")
	}
	contextUser := impart.GetCtxUser(ctx)
	if contextUser == nil || contextUser.ImpartWealthID == "" {
		return impart.NewError(impart.ErrBadRequest, "context user not found")
	}
	if contextUser.Admin {
		errorString := "Admin user doesn't have the permission"
		ps.Logger().Error(errorString, zap.Any("error", errorString))
		return impart.NewError(impart.ErrUnauthorized, errorString)
	}

	userToDelete, err := ps.profileStore.GetUser(ctx, impartWealthID)
	if err != nil {
		return impart.NewError(impart.ErrUnauthorized, fmt.Sprintf("couldn't find profile for impartWealthID %s", impartWealthID))
	}

	// admin removed- APP-144
	if contextUser.ImpartWealthID != userToDelete.ImpartWealthID {
		ps.Logger().Info("request to delete a user failed validation", zap.String("deleteUser", userToDelete.ImpartWealthID),
			zap.String("contextUser", contextUser.ImpartWealthID))

		return impart.NewError(impart.ErrUnauthorized, "user is not authorized")
	}
	if userToDelete.Blocked {
		return impart.NewError(impart.ErrBadRequest, "Cannot delete blocked user.")
	}
	err = ps.profileStore.DeleteUserProfile(ctx, deleteUser, hardDelete)
	if err != nil {
		return impart.NewError(impart.ErrBadRequest, "User delete failed.")
	}
	return nil
}

func (ps *profileService) NewProfile(ctx context.Context, p models.Profile, apiVersion string) (models.Profile, impart.Error) {
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
		ps.Logger().Debug("Unable to locate device token",
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

	if impartErr := ps.validateNewProfile(ctx, p, apiVersion); impartErr != nil {
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
	dbUser.HiveUpdatedAt = impart.CurrentUTC()

	// if err != nil {
	// 	ps.Logger().Error("Token Sync Endpoint error", zap.Any("Error", err), zap.Any("contextUser", ctxUser), zap.Any("inputProfile", p))
	// }
	// dbUser.AwsSNSAppArn = endpointARN
	// hide this : end

	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	background := impart.GetAvatharBackground()
	backgroundindex := rand.Intn(len(background))
	dbUser.AvatarBackground = background[backgroundindex]

	letter := impart.GetAvatharLetters()
	letterindex := rand.Intn(len(letter))
	dbUser.AvatarLetter = letter[letterindex]
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

		//create user device
		userDevice, err := ps.CreateUserDevice(ctx, dbUser, p.UserDevices[0].UserDeviceToDBModel())
		if err != nil {
			impartErr := impart.NewError(impart.ErrBadRequest, fmt.Sprintf("unable to add/update the device information %v", err))
			ps.Logger().Error(impartErr.Error())
		}
		out.UserDevices = append(out.UserDevices, userDevice)

		// check the device id exists
		if p.UserDevices[0].DeviceToken != "" {
			// map for notification
			var isAdmin bool
			if ctxUser != nil && ctxUser.Admin {
				isAdmin = true
			} else {
				isAdmin = false
			}
			err = ps.MapDeviceForNotification(ctx, userDevice, isAdmin)
			if err != nil {
				impartErr := impart.NewError(impart.ErrBadRequest, fmt.Sprintf("an error occured in update mapping for notification %v", err))
				ps.Logger().Error(impartErr.Error())
			}
		}
	}
	// // Adding User to MailChimp
	mailChimpParams := &members.NewParams{
		EmailAddress: dbUser.Email,
		Status:       members.StatusSubscribed,
	}
	cfg, _ := config.GetImpart()
	ps.Logger().Info("Mailcimp -", zap.Any("MailchimpApikey", cfg.MailchimpApikey),
		zap.Any("MailchimpAudienceId", cfg.MailchimpAudienceId))
	_, err = members.New(cfg.MailchimpAudienceId, mailChimpParams)
	if err != nil {
		impartErr := impart.NewError(impart.ErrBadRequest, fmt.Sprintf("User is not  added to the mailchimp %v", err))
		ps.Logger().Error(impartErr.Error())
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

func (s *profileService) GetHive(ctx context.Context, hiveID uint64) (*dbmodels.Hive, impart.Error) {
	hive, err := dbmodels.Hives(
		dbmodels.HiveWhere.HiveID.EQ(hiveID)).One(ctx, s.db)
	if err != nil {
		return nil, impart.NewError(impart.ErrUnknown, "unable to get Hive of hiveId")
	}

	return hive, nil
}

func (ps *profileService) UpdateReadCommunity(ctx context.Context, p models.UpdateReadCommunity, impartID string) impart.Error {
	existingDBUser, err := ps.profileStore.GetUser(ctx, impartID)
	if err != nil {
		return impart.NewError(impart.ErrUnknown, "unable to fetch existing user from dB")
	}
	existingDBProfile := existingDBUser.R.ImpartWealthProfile
	ps.Logger().Debug("Checking Updated Profile",
		zap.Any("existingDBUser", *existingDBUser),
		zap.Any("existingDBProfile", *existingDBProfile),
		zap.Any("updated", p))
	if p.IsUpdate != existingDBProfile.IsUpdateReadCommunity {
		existingDBProfile.IsUpdateReadCommunity = p.IsUpdate
	}
	err = ps.profileStore.UpdateProfile(ctx, existingDBUser, existingDBProfile)
	if err != nil {
		ps.Logger().Error("Upadte Read Community  user requset failed", zap.String("UpadteReadCommunity", impartID))
		return impart.NewError(err, "Upadte Read Community failed")
	}
	return nil
}

func (ps *profileService) DeleteUserByAdmin(ctx context.Context, hardDelete bool, deleteUser models.DeleteUserInput) impart.Error {
	if strings.TrimSpace(deleteUser.ImpartWealthID) == "" {
		return impart.NewError(impart.ErrBadRequest, "impartWealthID is empty.")
	}
	contextUser := impart.GetCtxUser(ctx)
	if contextUser == nil || contextUser.ImpartWealthID == "" {
		return impart.NewError(impart.ErrBadRequest, "context user not found.")
	}
	if !contextUser.SuperAdmin {
		errorString := "Current user does not have the permission."
		ps.Logger().Error(errorString, zap.Any("error", errorString))
		return impart.NewError(impart.ErrUnauthorized, errorString)
	}
	userToDelete, err := ps.profileStore.GetUser(ctx, deleteUser.ImpartWealthID)
	if err != nil {
		return impart.NewError(impart.ErrUnauthorized, fmt.Sprintf("could not find profile for impartWealthID %s", deleteUser.ImpartWealthID))
	}

	if userToDelete.Blocked {
		errorString := "Cannot delete  blocked user."
		ps.Logger().Error(errorString, zap.Any("error", errorString))
		return impart.NewError(impart.ErrUnauthorized, errorString)
	}
	if userToDelete.ImpartWealthID == contextUser.ImpartWealthID {
		errorString := "You cannot delete logged in user."
		ps.Logger().Error(errorString, zap.Any("error", errorString))
		return impart.NewError(impart.ErrUnauthorized, errorString)
	}
	imperr := ps.profileStore.DeleteUserProfile(ctx, deleteUser, hardDelete)
	if imperr != nil {
		return impart.NewError(impart.ErrBadRequest, "User delete failed.")
	}
	return nil
}

func (ps *profileService) CreatePlaidProfile(ctx context.Context, plaid models.PlaidInput) (models.PlaidInput, impart.Error) {
	ctxUser := impart.GetCtxUser(ctx)
	dbUser, err := ps.profileStore.GetUser(ctx, ctxUser.ImpartWealthID)
	if err != nil {
		ps.Error(err)
		return models.PlaidInput{}, impart.NewError(impart.ErrBadRequest, "Unable to find profile.")
	}
	if dbUser.ImpartWealthID != plaid.ImpartWealthID {
		return models.PlaidInput{}, impart.NewError(impart.ErrUnauthorized, "unable to edit a profile that's not yours.")
	}
	dbUser.PlaidAccessToken = null.StringFrom(plaid.PlaidAccessToken)
	err = ps.profileStore.UpdateProfile(ctx, dbUser, nil)
	if err != nil {
		ps.Error(err)
		return models.PlaidInput{}, impart.NewError(impart.ErrBadRequest, "Unable to update profile.")
	}

	return models.PlaidInput{}, nil
}

func (ps *profileService) GetWeeklyNotification(ctx context.Context) {
	impart.NotifyWeeklyActivity(ps.db, ps.Logger())
}

func (ps *profileService) GetWeeklyMostPopularNotification(ctx context.Context) {
	impart.NotifyWeeklyMostPopularPost(ps.db, ps.Logger())
}

func (ps *profileService) GetHiveNotification(ctx context.Context) error {
	err := ps.profileStore.GetHiveNotification(ctx)
	if err != nil {
		return err
	}
	return nil
}
