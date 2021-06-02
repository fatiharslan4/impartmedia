package profile

import (
	"context"
	"fmt"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"go.uber.org/zap"
)

func (ps *profileService) CreateUserDevice(ctx context.Context, ud *dbmodels.UserDevice) (models.UserDevice, impart.Error) {
	contextUser := impart.GetCtxUser(ctx)
	if contextUser == nil || contextUser.ImpartWealthID == "" {
		return models.UserDevice{}, impart.NewError(impart.ErrBadRequest, "context user not found")
	}

	// check the device details already exists in table
	// then dont insert to table
	exists, err := ps.profileStore.GetUserDevice(ctx, "", contextUser.ImpartWealthID, ud.DeviceID)
	if err != nil && err != impart.ErrNotFound {
		return models.UserDevice{}, impart.NewError(impart.ErrBadRequest, "error to find user device")
	}

	// if the entry for user is not exists
	if exists == nil {
		ud.ImpartWealthID = contextUser.ImpartWealthID
		response, err := ps.profileStore.CreateUserDevice(ctx, ud)
		if err != nil && err != impart.ErrNotFound {
			return models.UserDevice{}, impart.NewError(impart.ErrBadRequest, "error to create user device")
		}
		ud = response
	} else {
		ud = exists
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

/**
 * map the device id for notification
 *
 * check the user have configuration save for notification
 * if yes
 *		check it is true, then new config for notification map for true, else false
 * if No
 *		inseet with true
 */
func (ps *profileService) MapDeviceForNotification(ctx context.Context, ud models.UserDevice) impart.Error {
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
	// mapErr := ps.profileStore.DeleteUserNotificationMappData(ctx, "", ud.DeviceID, "")

	// check the same device is actived for some other users, then update the status into false
	// mapErr := ps.profileStore.UpdateExistingNotificationMappData(ctx, ud.ImpartWealthID, ud.DeviceID, "", true)
	mapErr := ps.profileStore.UpdateExistingNotificationMappData(models.MapArgumentInput{
		Ctx:            ctx,
		ImpartWealthID: ud.ImpartWealthID,
		DeviceID:       ud.DeviceID,
		Negate:         true,
	}, false)
	if mapErr != nil {
		errorString := fmt.Sprintf("error occured during update existing %s device id", ud.DeviceID)
		ps.Logger().Error(errorString, zap.Any("error", mapErr))
		return impart.NewError(impart.ErrBadRequest, errorString)
	}

	// check the entry already exists in db table
	// if yes then update the status only,
	// else add new
	exists, existsErr := ps.profileStore.GetUserNotificationMappData(models.MapArgumentInput{
		Ctx:            ctx,
		ImpartWealthID: ud.ImpartWealthID,
		DeviceID:       ud.DeviceID,
	})
	if existsErr != nil {
		errorString := fmt.Sprintf("unable to fetch the existing mapped data %s device id", ud.DeviceID)
		ps.Logger().Error(errorString, zap.Any("error", mapErr))
		return impart.NewError(impart.ErrBadRequest, errorString)
	}

	//there us no mapp entry exists , insert new entry
	if exists == nil {
		// from here, this device id should be sync with sns
		arn, err := ps.notificationService.SyncTokenEndpoint(ctx, ud.DeviceID, "")
		if err != nil {
			ps.Logger().Error("Token Sync Endpoint error", zap.Any("Error", err), zap.Any("contextUser", impart.GetCtxUser(ctx)))
		}

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
			DeviceID:       ud.DeviceID,
		}, notifyStatus)

		if mapErr != nil {
			errorString := fmt.Sprintf("error occure during delete existing %s device id", ud.DeviceID)
			ps.Logger().Error(errorString, zap.Any("error", mapErr))
			return impart.NewError(impart.ErrBadRequest, errorString)
		}

	}
	if mapErr != nil {
		errorString := fmt.Sprintf("unable to add %s device id", ud.DeviceID)
		ps.Logger().Error(errorString, zap.Any("error", mapErr))
		return impart.NewError(impart.ErrBadRequest, errorString)
	}

	return nil
}
