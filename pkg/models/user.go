package models

import (
	"context"
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

type UserConfigurations struct {
	ConfigID           uint
	ImpartWealthID     string `json:"impartWealthId"`
	NotificationStatus bool   `json:"notificationStatus"`
}

func (d UserDevice) UserDeviceToDBModel() *dbmodels.UserDevice {
	out := &dbmodels.UserDevice{
		Token:          d.Token,
		ImpartWealthID: d.ImpartWealthID,
		DeviceID:       d.DeviceID,
		AppVersion:     d.AppVersion,
		DeviceName:     d.DeviceName,
		DeviceVersion:  d.DeviceVersion,
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
		DeviceVersion:  d.DeviceVersion,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
		DeletedAt:      d.DeletedAt,
	}

	return out
}

func UserConfigurationFromDBModel(d *dbmodels.UserConfiguration) UserConfigurations {
	out := UserConfigurations{
		ConfigID:           d.ConfigID,
		ImpartWealthID:     d.ImpartWealthID,
		NotificationStatus: d.NotificationStatus,
	}

	return out
}

func (uc UserConfigurations) UserConfigurationTODBModel() *dbmodels.UserConfiguration {
	out := &dbmodels.UserConfiguration{
		ConfigID:           uc.ConfigID,
		ImpartWealthID:     uc.ImpartWealthID,
		NotificationStatus: uc.NotificationStatus,
	}

	return out
}

// filter input
type MapArgumentInput struct {
	Ctx            context.Context
	ImpartWealthID string
	DeviceID       string
	DeviceToken    string
	Negate         bool
}

// user Notification
type UserGlobalConfigInput struct {
	DeviceToken    string `json:"deviceToken,omitempty"`
	Status         bool   `json:"status"`
	ImpartWealthID string `json:"impartWealthID,omitempty"`
	Type           string `json:"type"`
}

type BlockUserInput struct {
	ImpartWealthID string `json:"impartWealthID,omitempty"`
	ScreenName     string `json:"screenName,omitempty"`
	Status         string `json:"status,omitempty" default:"block"`
}
