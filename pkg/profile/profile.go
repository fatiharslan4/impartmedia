package profile

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	profile_data "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/leebenson/conform"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
)

type Service interface {
	NewProfile(contextAuthID string, p models.Profile) (models.Profile, impart.Error)
	GetProfile(GetProfileInput) (models.Profile, impart.Error)
	UpdateProfile(contextAuthID string, p models.Profile) (models.Profile, impart.Error)
	DeleteProfile(contextAuthID, impartWealthID string) impart.Error
	ScreenNameExists(screenName string) bool

	WhiteListSearch(impartWealthId, email, screenName string) (models.WhiteListProfile, impart.Error)
	UpdateWhitelistScreenName(impartWealthId, screenName string) impart.Error
	CreateWhitelistEntry(profile models.WhiteListProfile) impart.Error
	ValidateSchema(document gojsonschema.JSONLoader) impart.Error
	Logger() *zap.Logger
}

func New(logger *zap.SugaredLogger, dal profile_data.Store, ns impart.NotificationService, schema gojsonschema.JSONLoader, stage string) Service {
	return &profileService{
		stage:               stage,
		SugaredLogger:       logger,
		db:                  dal,
		notificationService: ns,
		schemaValidator:     schema,
	}
}

type profileService struct {
	stage string
	*zap.SugaredLogger
	db                  profile_data.Store
	schemaValidator     gojsonschema.JSONLoader
	notificationService impart.NotificationService
}

func (ps *profileService) ScreenNameExists(screenName string) bool {
	impartId, err := ps.db.GetImpartIdFromScreenName(screenName)
	if err == nil && impartId != "" {
		return true
	}
	return false
}

func (ps *profileService) Logger() *zap.Logger {
	return ps.Desugar()
}

func (ps *profileService) DeleteProfile(contextAuthID, impartWealthID string) impart.Error {
	if strings.TrimSpace(impartWealthID) == "" {
		return impart.NewError(impart.ErrBadRequest, "impartWealthID is empty")
	}
	p, err := ps.db.GetProfile(impartWealthID, true)
	if err != nil {
		return impart.NewError(err, fmt.Sprintf("couldn't find profile for impartWealthID %s", impartWealthID))
	}

	// If the user is not deleting their own profile, and the user is not an admin, error out.
	if p.AuthenticationID != contextAuthID {
		contextUserProfile, err := ps.db.GetProfileFromAuthId(contextAuthID, false)
		if err != nil {
			return impart.NewError(err, "unable to get context profile after mismatch "+
				"on requested delete profile autID")
		}
		if !contextUserProfile.Attributes.Admin {
			return impart.NewError(impart.ErrUnauthorized, "user is not authorized")
		}
	}

	err = ps.db.DeleteProfile(impartWealthID)
	if err != nil {
		return impart.NewError(err, "unable to retrieve profile")
	}

	return nil

}

func (ps *profileService) NewProfile(contextAuthID string, p models.Profile) (models.Profile, impart.Error) {
	var empty models.Profile
	var err error

	err = conform.Strings(&p)
	if err != nil {
		return empty, impart.NewError(err, "unable to conform profile")
	}

	if contextAuthID == "" {
		return empty, impart.NewError(impart.ErrBadRequest, "Unable to locate authenticationID")
	}
	p.AuthenticationID = contextAuthID

	// Populate whitelist entries if they exist
	whitelistEntry, impartErr := ps.WhiteListSearch(p.ImpartWealthID, "", "")
	if impartErr != nil {
		if impartErr.Err() == impart.ErrNotFound {
			p.SurveyResponses = models.SurveyResponses{}
		} else {
			return empty, impartErr
		}
	} else {
		p.SurveyResponses = whitelistEntry.SurveyResponses
	}

	// If no screenName is provided, use screenName from whitelist Entry.
	if p.ScreenName == "" && whitelistEntry.ScreenName != "" {
		p.ScreenName = whitelistEntry.ScreenName
	}

	if impartErr := ps.validateNewProfile(p); impartErr != nil {
		return empty, impartErr
	}

	p.CreatedDate = impart.CurrentUTC()
	p.UpdatedDate = impart.CurrentUTC()
	p.Attributes.UpdatedDate = impart.CurrentUTC()
	p.SurveyResponses.ImportTimestamp = impart.CurrentUTC()

	if p.Attributes.Admin && ps.stage == "prod" {
		p.Attributes.Admin = false
	}

	p.NotificationProfile, err = ps.SubscribeDeviceToken(p)
	if err != nil {
		return empty, impart.NewError(err, "error syncing profile device token and hive subscriptions")
	}

	created, err := ps.db.CreateProfile(p)
	if err != nil {
		ps.Error(err)
		return empty, impart.NewError(err, "unable to create profile")
	}

	//hides subscriptions from impart app temporarily
	//TODO: remove this
	created.NotificationProfile.Subscriptions = nil
	return created, nil
}

type GetProfileInput struct {
	ImpartWealthID, SearchAuthenticationID, SearchEmail, SearchScreenName, ContextAuthID string
}

func (ps *profileService) GetProfile(gpi GetProfileInput) (models.Profile, impart.Error) {
	var err error
	var p models.Profile
	var impartWealthID string

	getProfileFunc := func(id string) (models.Profile, impart.Error) {
		if p, err = ps.db.GetProfile(id, false); err != nil {
			return p, impart.NewError(err, "unable to retrieve profile")
		}
		if p.AuthenticationID != gpi.ContextAuthID {
			return models.Profile{}, impart.NewError(impart.ErrUnauthorized, fmt.Sprintf("authentication ID in token does not match for impart Id %s", id))
		}
		//hides subscriptions from impart app temporarily
		//TODO: remove this
		p.NotificationProfile.Subscriptions = nil
		return p, nil
	}

	if gpi.ImpartWealthID != "" {
		return getProfileFunc(gpi.ImpartWealthID)
	}

	if gpi.SearchAuthenticationID != "" {
		if impartWealthID, err = ps.db.GetImpartIdFromAuthId(gpi.SearchAuthenticationID); err != nil {
			return p, impart.NewError(err, "unable to find the profile for authenticationId "+gpi.SearchAuthenticationID)
		}
		return getProfileFunc(impartWealthID)
	}

	if gpi.SearchEmail != "" {
		if impartWealthID, err = ps.db.GetImpartIdFromEmail(gpi.SearchEmail); err != nil {
			return p, impart.NewError(err, "unable to find the profile for email"+gpi.SearchEmail)
		}
		return getProfileFunc(impartWealthID)
	}
	if gpi.SearchScreenName != "" {
		if impartWealthID, err = ps.db.GetImpartIdFromScreenName(gpi.SearchScreenName); err != nil {
			return p, impart.NewError(err, "unable to find the profile for screenName"+gpi.SearchScreenName)
		}
		return getProfileFunc(impartWealthID)
	}

	return p, impart.NewError(impart.ErrBadRequest, "no valid query parameters to return profile")
}

func (ps *profileService) UpdateProfile(contextAuthID string, p models.Profile) (models.Profile, impart.Error) {
	var originalProfile models.Profile
	var empty models.Profile
	var err error

	if err = conform.Strings(&p); err != nil {
		return models.Profile{}, impart.NewError(err, "invalid message format")
	}

	requestorProfile, err := ps.db.GetProfileFromAuthId(contextAuthID, false)
	if err != nil {
		return p, impart.NewError(err, fmt.Sprintf("unable to find requestors profile from authenticationId %s", contextAuthID))
	}

	if p.AuthenticationID != contextAuthID && !requestorProfile.Attributes.Admin {
		return models.Profile{}, impart.NewError(impart.ErrUnauthorized, fmt.Sprintf("user %s is not authorized to modify profile of %s: %s",
			requestorProfile.ScreenName, p.ScreenName, p.ImpartWealthID))
	}

	if originalProfile, err = ps.db.GetProfile(p.ImpartWealthID, true); err != nil {
		return empty, impart.NewError(err, fmt.Sprintf("error retrieving existing profile"))
	}

	ps.Logger().Debug("Checking Updated Profile",
		zap.Any("original", originalProfile),
		zap.Any("updated", p))

	if p.CreatedDate.IsZero() ||
		p.UpdatedDate.IsZero() ||
		p.CreatedDate != originalProfile.CreatedDate ||
		p.UpdatedDate != originalProfile.UpdatedDate {
		msg := "profile being updated appears to be incorrect - critical properties do not match the profile being updated."
		ps.Logger().Error(msg, zap.Any("updatedProfile", p))
		return empty, impart.NewError(impart.ErrBadRequest, msg)
	}

	err = ps.UpdateProfileSync(originalProfile, &p)

	p.UpdatedDate = impart.CurrentUTC()
	p, err = ps.db.UpdateProfile(contextAuthID, p)
	if err != nil {
		return empty, impart.NewError(err, fmt.Sprintf("Unable to update profile %s", originalProfile.ImpartWealthID))
	}
	//hides subscriptions from impart app temporarily
	//TODO: remove this
	p.NotificationProfile.Subscriptions = nil

	return p, nil
}

// UpdateProfileSync takes a pointer to the profile that is being updated makes sure the incoming changes has all the appropriate additional touches
func (ps *profileService) UpdateProfileSync(originalProfile models.Profile, updatedProfile *models.Profile) (err error) {
	if err = ps.CheckNotificationProfileChanged(originalProfile, updatedProfile); err != nil {
		return err
	}

	return err
}

func (ps *profileService) CheckNotificationProfileChanged(currentProfile models.Profile, updatedProfile *models.Profile) error {

	///unhides subscriptions from impart app temporarily
	//TODO: remove this
	if strings.TrimSpace(updatedProfile.NotificationProfile.DeviceToken) != "" {
		updatedProfile.NotificationProfile.Subscriptions = currentProfile.NotificationProfile.Subscriptions
	}

	currentNotificationProfile := currentProfile.NotificationProfile
	updatedNotificationProfile := updatedProfile.NotificationProfile

	if cmp.Equal(currentNotificationProfile, updatedNotificationProfile, cmpopts.EquateEmpty(),
		cmpopts.SortSlices(
			func(i, j models.Subscription) bool {
				return i.Name < j.Name
			})) {
		return nil
	}

	//Check for bad updates with missing subs
	if currentNotificationProfile.DeviceToken == updatedNotificationProfile.DeviceToken && len(currentNotificationProfile.Subscriptions) != len(updatedNotificationProfile.Subscriptions) {
		ps.Logger().Warn("received a changed profile where device tokens were the same, but subscriptions were not - not updating anything")
		//De-refs the pointers to set the new updated profile back to the current profile
		updatedNotificationProfile = currentNotificationProfile
		return nil
	}

	// Check if we're removing notification profile
	if currentNotificationProfile.DeviceToken != "" && updatedNotificationProfile.DeviceToken == "" {
		for _, subscription := range currentNotificationProfile.Subscriptions {
			err := ps.notificationService.UnsubscribeTopic(subscription.SubscriptionARN)
			// Swallows error and moves on.
			if err != nil {
				ps.Logger().Error("error unsubscribing to topic", zap.Error(err), zap.Any("subscription", subscription))
			}
		}
		return nil
	}

	// Check if we're changing device token
	if currentNotificationProfile.DeviceToken != updatedNotificationProfile.DeviceToken {
		//endpointARN, err := ps.notificationService.SyncTokenEndpoint(updatedNotificationProfile.DeviceToken, updatedNotificationProfile.AWSPlatformEndpointARN)
		//if err != nil {
		//	return err
		//}
		//
		//var memberships models.HiveMemberships
		//if len(updatedProfile.Attributes.HiveMemberships) < 0 {
		//	memberships = updatedProfile.Attributes.HiveMemberships
		//} else {
		//	memberships = currentProfile.Attributes.HiveMemberships
		//}
		////Add hive topics
		//for _, h := range memberships{
		//	hive, err := ps.db.GetHive(h.HiveID, false)
		//	if err != nil {
		//		ps.Logger().Error("error getting hive from DB", zap.Error(err))
		//	}
		//	subscription, err := ps.notificationService.SubscribeTopic(endpointARN, hive.PinnedPostNotificationTopicARN)
		//	if err != nil {
		//		ps.Logger().Error("error subscribing to topic", zap.Error(err))
		//	}
		//	updatedNotificationProfile.Subscriptions = append(updatedNotificationProfile.Subscriptions, models.Subscription{Name: hive.PinnedPostNotificationTopicARN, SubscriptionARN: subscription})
		//}
		//
		//updatedNotificationProfile.AWSPlatformEndpointARN = endpointARN

		//updatedProfile.NotificationProfile = updatedNotificationProfile
		var err error
		updatedProfile.NotificationProfile, err = ps.SubscribeDeviceToken(*updatedProfile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ps *profileService) SubscribeDeviceToken(profile models.Profile) (models.NotificationProfile, error) {
	notificationProfile := profile.NotificationProfile

	if strings.TrimSpace(notificationProfile.DeviceToken) == "" {
		return models.NotificationProfile{}, nil
	}

	endpointARN, err := ps.notificationService.SyncTokenEndpoint(notificationProfile.DeviceToken, notificationProfile.AWSPlatformEndpointARN)
	if err != nil {
		return models.NotificationProfile{}, err
	}

	notificationProfile.AWSPlatformEndpointARN = endpointARN

	for _, h := range profile.Attributes.HiveMemberships {
		hive, err := ps.db.GetHive(h.HiveID, false)
		if err != nil {
			ps.Logger().Error("error getting hive from DB", zap.Error(err))
		}
		subscription, err := ps.notificationService.SubscribeTopic(endpointARN, hive.PinnedPostNotificationTopicARN)
		if err != nil {
			ps.Logger().Error("error subscribing to topic", zap.Error(err))
		}
		notificationProfile.Subscriptions = append(notificationProfile.Subscriptions, models.Subscription{Name: hive.PinnedPostNotificationTopicARN, SubscriptionARN: subscription})
	}

	return notificationProfile, nil
}
