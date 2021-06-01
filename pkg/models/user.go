package models

import (
	"time"

	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
)

type UserDevice struct {
	Token          string    `json:"token,omitempty"`
	ImpartWealthID string    `json:"impartWealthId,omitempty" conform:"trim"`
	DeviceID       string    `json:"deviceId"`
	AppVersion     string    `json:"appVersion"`
	DeviceName     string    `json:"deviceName"`
	DeviceVersion  string    `json:"deviceVersion"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
	DeletedAt      null.Time `json:"deleted_at,omitempty"`
}

func (d UserDevice) ToDBModel() *dbmodels.UserDevice {
	out := &dbmodels.UserDevice{
		Token:          []byte(d.Token),
		ImpartWealthID: d.ImpartWealthID,
		DeviceID:       d.DeviceID,
		AppVersion:     d.AppVersion,
		DeviceName:     d.DeviceName,
		DeviceVersion:  null.StringFrom(d.DeviceVersion),
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
		DeletedAt:      d.DeletedAt,
	}

	return out
}

func UserDeviceFromDBModel(d *dbmodels.UserDevice) UserDevice {
	out := UserDevice{
		Token:          string(d.Token),
		ImpartWealthID: d.ImpartWealthID,
		DeviceID:       d.DeviceID,
		AppVersion:     d.AppVersion,
		DeviceName:     d.DeviceName,
		DeviceVersion:  d.DeviceVersion.String,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
		DeletedAt:      d.DeletedAt,
	}

	return out
}
