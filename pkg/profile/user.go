package profile

import (
	"context"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
)

func (ps *profileService) AddUserDevice(ctx context.Context, ud *dbmodels.UserDevice) (models.UserDevice, impart.Error) {
	contextUser := impart.GetCtxUser(ctx)
	if contextUser == nil || contextUser.ImpartWealthID == "" {
		return models.UserDevice{}, impart.NewError(impart.ErrBadRequest, "context user not found")
	}

	// check the device details already exists in table
	// then dont insert to table
	exists, err := ps.profileStore.GetUserDevice(ctx, []byte(ud.DeviceID), contextUser.ImpartWealthID)
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
