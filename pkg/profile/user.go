package profile

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/beeker1121/mailchimp-go/lists/members"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/zap"
)

func (ps *profileService) GetUserDevice(ctx context.Context, token string, impartID string, deviceToken string) (models.UserDevice, error) {
	device, err := ps.profileStore.GetUserDevice(ctx, token, impartID, deviceToken)
	if err != nil {
		errorString := fmt.Sprintf("error occured during update existing %s device token", deviceToken)
		ps.Logger().Error(errorString, zap.Any("error", err))
		return models.UserDevice{}, impart.NewError(impart.ErrBadRequest, errorString)
	}

	return models.UserDeviceFromDBModel(device), nil
}

func (ps *profileService) CreateUserDevice(ctx context.Context, user *dbmodels.User, ud *dbmodels.UserDevice) (models.UserDevice, impart.Error) {
	var contextUser *dbmodels.User
	if user == nil {
		contextUser = impart.GetCtxUser(ctx)
	} else {
		contextUser = user
	}

	if contextUser == nil || contextUser.ImpartWealthID == "" {
		return models.UserDevice{}, impart.NewError(impart.ErrBadRequest, "context user not found")
	}

	deviceToken := "__NILL__"
	if ud.DeviceToken != "" {
		deviceToken = ud.DeviceToken
	}

	// check the device details already exists in table
	// then dont insert to table
	exists, err := ps.profileStore.GetUserDevice(ctx, "", contextUser.ImpartWealthID, deviceToken)
	if err != nil && err != impart.ErrNotFound {
		return models.UserDevice{}, impart.NewError(impart.ErrBadRequest, "error to find user device")
	}

	// if the entry for user is not exists
	if exists == nil {
		ud.ImpartWealthID = contextUser.ImpartWealthID
		response, err := ps.profileStore.CreateUserDevice(ctx, ud)
		if err != nil && err != impart.ErrNotFound {
			return models.UserDevice{}, impart.NewError(impart.ErrBadRequest, fmt.Sprintf("error to create user device %v", err))
		}
		ud = response
	} else {
		exists.AppVersion = ud.AppVersion
		exists.DeviceID = ud.DeviceID
		exists.DeviceName = ud.DeviceName
		exists.DeviceVersion = ud.DeviceVersion
		err = ps.profileStore.UpdateDevice(ctx, exists)
		if err != nil && err != impart.ErrNotFound {
			return models.UserDevice{}, impart.NewError(impart.ErrBadRequest, fmt.Sprintf("error to create user device %v", err))
		}
		ud = exists
	}

	userToUpdate, err := ps.profileStore.GetUser(ctx, contextUser.ImpartWealthID)
	if err == nil {
		existingDBProfile := userToUpdate.R.ImpartWealthProfile
		currTime := time.Now().In(boil.GetLocation())
		userToUpdate.LastloginAt = null.TimeFrom(currTime)
		err = ps.profileStore.UpdateProfile(ctx, userToUpdate, existingDBProfile)
		if err != nil {
			ps.Logger().Error("Update user last login requset failed", zap.String("Update", userToUpdate.ImpartWealthID),
				zap.String("contextUser", contextUser.ImpartWealthID))
		}
	}

	return models.UserDeviceFromDBModel(ud), nil
}

func (ps *profileService) GetUserConfigurations(ctx context.Context, impartWealthID string) (models.UserConfigurations, impart.Error) {
	config, err := ps.profileStore.GetUserConfigurations(ctx, impartWealthID)
	if err != nil && err != impart.ErrNotFound {
		return models.UserConfigurations{}, impart.NewError(impart.ErrBadRequest, "error to get user configurations")
	}

	if config == nil {
		return models.UserConfigurations{}, nil
	}
	return models.UserConfigurationFromDBModel(config), nil
}

// map the device id for notification
// check the user have configuration save for notification
//  if yes
//		check it is true, then new config for notification map for true, else false
//  if No
//		insert with true
func (ps *profileService) MapDeviceForNotification(ctx context.Context, ud models.UserDevice, isAdmin bool) impart.Error {
	var notifyStatus bool
	userConfig, err := ps.GetUserConfigurations(ctx, ud.ImpartWealthID)
	if err != nil {
		return impart.NewError(impart.ErrBadRequest, "unable to read user configurations")
	}
	//check the user have a global configuration for notification enable
	if (models.UserConfigurations{} != userConfig) {
		if userConfig.NotificationStatus {
			notifyStatus = true
		}
	} else {
		notifyStatus = true
	}

	// check the same device is accessed for another user, then have to remove that
	// remove the entries and insert new entry
	// delete all the entries with the same device id
	// mapErr := ps.profileStore.DeleteUserNotificationMappData(ctx, "", ud.DevticeToken, "")

	// check the same device is actived for some other users, then update the status into false
	// mapErr := ps.profileStore.UpdateExistingNotificationMappData(ctx, ud.ImpartWealthID, ud.DeviceToken, "", true)
	mapErr := ps.profileStore.UpdateExistingNotificationMappData(models.MapArgumentInput{
		Ctx:            ctx,
		ImpartWealthID: ud.ImpartWealthID,
		DeviceToken:    ud.DeviceToken,
		Negate:         true,
	}, false)
	if mapErr != nil && err != sql.ErrNoRows {
		errorString := fmt.Sprintf("error occured during update existing %s device id", ud.DeviceToken)
		ps.Logger().Error(errorString, zap.Any("error", mapErr))
		return impart.NewError(impart.ErrBadRequest, errorString)
	}

	// check the entry already exists in db table
	// if yes then update the status only,
	// else add new
	exists, existsErr := ps.profileStore.GetUserNotificationMappData(models.MapArgumentInput{
		Ctx:            ctx,
		ImpartWealthID: ud.ImpartWealthID,
		DeviceToken:    ud.DeviceToken,
	})
	if existsErr != nil {
		errorString := fmt.Sprintf("unable to fetch the existing mapped data %s device id", ud.DeviceToken)
		ps.Logger().Error(errorString, zap.Any("error", mapErr))
		return impart.NewError(impart.ErrBadRequest, errorString)
	}

	// from here, this device id should be sync with sns
	arn, nErr := ps.notificationService.SyncTokenEndpoint(ctx, ud.DeviceToken, "")
	if nErr != nil {
		ps.Logger().Error("Token Sync Endpoint error",
			zap.Any("Error", nErr),
			zap.Any("Device", ud),
		)
	}

	//subscribe unsubsribe to topic
	var hiveId uint64
	ctxUser, userErr := ps.profileStore.GetUser(ctx, ud.ImpartWealthID)
	if userErr == nil {
		// user  active
		if ctxUser.R.MemberHiveHives != nil {
			for _, h := range ctxUser.R.MemberHiveHives {
				hiveId = h.HiveID
			}
		}
	}
	hiveData, err := ps.GetHive(ctx, hiveId)

	if err != nil {
		return impart.NewError(impart.ErrBadRequest, "unable to read user configurations")
	}
	if notifyStatus {
		if !isAdmin {
			if hiveData != nil && hiveData.NotificationTopicArn.String != "" {
				ps.notificationService.SubscribeTopic(ctx, ud.ImpartWealthID, hiveData.NotificationTopicArn.String, arn)
			}
		} else {
			if hiveData != nil && hiveData.NotificationTopicArn.String != "" {
				ps.notificationService.UnsubscribeTopicForDevice(ctx, ud.ImpartWealthID, hiveData.NotificationTopicArn.String, arn)
			}
		}
	} else {
		if hiveData != nil && hiveData.NotificationTopicArn.String != "" {
			ps.notificationService.UnsubscribeTopicForDevice(ctx, ud.ImpartWealthID, hiveData.NotificationTopicArn.String, arn)
		}

	}

	//there is no mapp entry exists , insert new entry
	if exists == nil {

		_, mapErr = ps.profileStore.CreateUserNotificationMappData(ctx, &dbmodels.NotificationDeviceMapping{
			ImpartWealthID: ud.ImpartWealthID,
			UserDeviceID:   ud.Token,
			NotifyStatus:   notifyStatus,
			NotifyArn:      arn,
		})

	} else {
		//update the existing map data status
		mapErr := ps.profileStore.UpdateExistingNotificationMappData(models.MapArgumentInput{
			Ctx:            ctx,
			ImpartWealthID: ud.ImpartWealthID,
			DeviceToken:    ud.DeviceToken,
		}, notifyStatus)

		if mapErr != nil {
			errorString := fmt.Sprintf("error occure during delete existing %s device token", ud.DeviceName)
			ps.Logger().Error(errorString, zap.Any("error", mapErr))
			return impart.NewError(impart.ErrBadRequest, errorString)
		}

	}
	if mapErr != nil {
		errorString := fmt.Sprintf("unable to add %s device token", ud.DeviceToken)
		ps.Logger().Error(errorString, zap.Any("error", mapErr))
		return impart.NewError(impart.ErrBadRequest, errorString)
	}

	return nil
}

// Save user configuration
func (ps *profileService) ModifyUserConfigurations(ctx context.Context, conf models.UserConfigurations) (models.UserConfigurations, impart.Error) {
	var configuration *dbmodels.UserConfiguration
	var err error
	// check the user config alreay exists
	// then update, else insert
	configuration, err = ps.profileStore.GetUserConfigurations(ctx, conf.ImpartWealthID)
	if err != nil && err != impart.ErrNotFound {
		errorString := "unable to get the user configuration"
		ps.Logger().Error(errorString, zap.Any("error", err))
		return models.UserConfigurations{}, impart.NewError(impart.ErrBadRequest, errorString)
	}
	// if the entry exists, then update with latest
	if configuration != nil {
		configuration.NotificationStatus = conf.NotificationStatus
		configuration, err = ps.profileStore.EditUserConfigurations(ctx, configuration)
	} else {
		configuration, err = ps.profileStore.CreateUserConfigurations(ctx, conf.UserConfigurationTODBModel())
	}
	if err != nil {
		errorString := "unable to add/update user configuration"
		ps.Logger().Error(errorString, zap.Any("error", err))
		return models.UserConfigurations{}, impart.NewError(impart.ErrBadRequest, errorString)
	}

	return models.UserConfigurationFromDBModel(configuration), nil
}

// Update Existing Notification Mapp Data
// Which will upodate the notification mapp status into true/false
func (ps *profileService) UpdateExistingNotificationMappData(input models.MapArgumentInput, status bool) impart.Error {
	err := ps.profileStore.UpdateExistingNotificationMappData(input, status)
	if err != nil {
		errorString := "unable to update notification map status"
		ps.Logger().Error(errorString, zap.Any("error", err))
		return impart.NewError(impart.ErrBadRequest, errorString)
	}
	return nil
}

// Block user
func (ps *profileService) BlockUser(ctx context.Context, impartID string, screenName string, status bool) impart.Error {
	ctxUser := impart.GetCtxUser(ctx)
	if !ctxUser.Admin {
		errorString := "current user doesn't have the permission"
		ps.Logger().Error(errorString, zap.Any("error", errorString))
		return impart.NewError(impart.ErrUnauthorized, errorString)
	}

	if impartID == "" && screenName == "" {
		errorString := "please provided user data to block"
		ps.Logger().Error(errorString, zap.Any("error", errorString))
		return impart.NewError(impart.ErrBadRequest, errorString)
	}

	//get user
	var dbUser *dbmodels.User
	var err error
	if impartID != "" {
		dbUser, err = ps.profileStore.GetUser(ctx, impartID)
	} else {
		dbUser, err = ps.profileStore.GetUserFromScreenName(ctx, screenName)
	}

	if err != nil {
		errorString := "unable to find user"
		return impart.NewError(impart.ErrBadRequest, errorString)
	}

	// cant block admin
	if dbUser.Admin {
		errorString := "cant't block admin user"
		return impart.NewError(impart.ErrBadRequest, errorString)
	}

	// block the user
	err = ps.profileStore.BlockUser(ctx, dbUser, status)
	if err != nil {
		errorString := fmt.Sprintf("%v", err)
		return impart.NewError(impart.ErrBadRequest, errorString)
	}
	exitingUserAnser := dbUser.R.ImpartWealthUserAnswers
	answerIds := make([]uint, len(exitingUserAnser))
	for i, a := range exitingUserAnser {
		answerIds[i] = a.AnswerID
	}
	hiveid := DefaultHiveId
	for _, h := range dbUser.R.MemberHiveHives {
		hiveid = h.HiveID
	}
	err = ps.profileStore.UpdateUserDemographic(ctx, answerIds, false)
	err = ps.profileStore.UpdateHiveUserDemographic(ctx, answerIds, false, hiveid)

	// // delete user from mailchimp
	// cfg, _ := config.GetImpart()
	err = members.Delete(impart.MailChimpAudienceID, dbUser.Email)
	if err != nil {
		ps.Logger().Error("Delete user requset failed in MailChimp", zap.String("blockUser", ctxUser.ImpartWealthID),
			zap.String("User", ctxUser.ImpartWealthID))
	}
	return nil
}

func (ps *profileService) UpdateDeviceToken(ctx context.Context, token string, deviceToken string) impart.Error {
	device, err := ps.profileStore.GetUserDevice(ctx, token, "", "")
	if err != nil {
		return impart.NewError(impart.ErrBadRequest, fmt.Sprintf("%v", err))
	}
	err = ps.profileStore.UpdateDeviceToken(ctx, device, deviceToken)
	if err != nil {
		return impart.NewError(impart.ErrBadRequest, fmt.Sprintf("%v", err))
	}
	return nil
}

func (ps *profileService) DeleteExceptUserDevice(ctx context.Context, impartID string, deviceToken string, refToken string) error {
	return ps.profileStore.DeleteExceptUserDevice(ctx, impartID, deviceToken, refToken)
}

func (ps *profileService) GetUsersDetails(ctx context.Context, gpi models.GetAdminInputs) ([]models.UserDetail, *models.NextPage, impart.Error) {
	result, nextPage, err := ps.profileStore.GetUsersDetails(ctx, gpi)
	if err != nil {
		ps.Logger().Error("Error in data fetching", zap.Error(err))
		return nil, nextPage, impart.NewError(impart.ErrUnknown, "unable to fetch the details")
	}
	return result, nextPage, nil
}

func (ps *profileService) GetPostDetails(ctx context.Context, gpi models.GetAdminInputs) ([]models.PostDetail, *models.NextPage, impart.Error) {
	result, nextPage, err := ps.profileStore.GetPostDetails(ctx, gpi)
	if err != nil {
		ps.Logger().Error("Error in data fetching", zap.Error(err))
		return nil, nextPage, impart.NewError(impart.ErrUnknown, "unable to fetch the details")
	}
	return result, nextPage, nil
}

func (ps *profileService) EditUserDetails(ctx context.Context, gpi models.WaitListUserInput) (string, impart.Error) {
	contextUser := impart.GetCtxUser(ctx)
	if contextUser == nil || contextUser.ImpartWealthID == "" {
		return "", impart.NewError(impart.ErrBadRequest, "context user not found.")
	}
	userToUpdate, err := ps.profileStore.GetUser(ctx, gpi.ImpartWealthID)
	if err != nil {
		ps.Logger().Error("Cannot Find the user", zap.Error(err))
		return "", impart.NewError(impart.ErrNotFound, "Cannot find the user")
	}
	if userToUpdate.Blocked {
		ps.Logger().Error("Blocked user", zap.Error(err))
		return "", impart.NewError(impart.ErrNotFound, "Blocked user")
	}
	msg, err0 := ps.profileStore.EditUserDetails(ctx, gpi)
	if err0 != nil {
		ps.Logger().Error("Error in adding waitlist", zap.Error(err))
		return msg, err0
	}
	return msg, nil
}

func (ps *profileService) GetHiveDetails(ctx context.Context, gpi models.GetAdminInputs) ([]map[string]interface{}, *models.NextPage, impart.Error) {
	// result, nextPage, err := ps.profileStore.GetHiveDetailsOld(ctx, gpi)
	result, nextPage, err := ps.profileStore.GetHiveDetails(ctx, gpi)
	if err != nil {
		ps.Logger().Error("Error in data fetching", zap.Error(err))
		return nil, nextPage, impart.NewError(impart.ErrUnknown, "unable to fetch the details")
	}
	return result, nextPage, nil
}

func (ps *profileService) GetFilterDetails(ctx context.Context) ([]byte, impart.Error) {
	result, err := ps.profileStore.GetFilterDetails(ctx)
	if err != nil {
		return nil, impart.NewError(impart.ErrNotFound, "Filter data fetching failed.")
	}
	return result, nil
}

func (ps *profileService) EditBulkUserDetails(ctx context.Context, userUpdates models.UserUpdate) (*models.UserUpdate, impart.Error) {
	userOutput := &models.UserUpdate{}
	if userUpdates.Action == "" {
		return nil, impart.NewError(impart.ErrBadRequest, "Incorrect input details")
	}
	if userUpdates.Action == "update" {
		if userUpdates.Type == "" {
			return nil, impart.NewError(impart.ErrBadRequest, "Incorrect input details")
		}
		if len(userUpdates.Users) == 0 {
			return nil, impart.NewError(impart.ErrBadRequest, "User details not found")
		}
		if userUpdates.Type == impart.AddToHive && userUpdates.HiveID == 0 {
			return nil, impart.NewError(impart.ErrBadRequest, "Missing hive details.")
		}
		userOutput := ps.profileStore.EditBulkUserDetails(ctx, userUpdates)
		ps.Logger().Info("bulk action proccess completed and return to route")
		return userOutput, nil
	} else if userUpdates.Action == "delete" {
		if len(userUpdates.Users) == 0 {
			return nil, impart.NewError(impart.ErrBadRequest, "User details not found")
		}
		userOutput = ps.profileStore.DeleteBulkUserDetails(ctx, userUpdates)
		return userOutput, nil
	}
	return userOutput, nil
}
