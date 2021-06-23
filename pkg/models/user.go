package models

import (
	"context"
	"time"

	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
)

type UserDevice struct {
	Token          string    `json:"token,omitempty"`
	ImpartWealthID string    `json:"impartWealthId,omitempty" conform:"trim"`
	DeviceID       string    `json:"deviceId"`
	DeviceToken    string    `json:"deviceToken,omitempty"`
	AppVersion     string    `json:"appVersion"`
	DeviceName     string    `json:"deviceName"`
	DeviceVersion  string    `json:"deviceVersion"`
	CreatedAt      time.Time `json:"createdAt,omitempty"`
	UpdatedAt      time.Time `json:"updatedAt,omitempty"`
}

type UserConfigurations struct {
	ConfigID           uint
	ImpartWealthID     string `json:"impartWealthId"`
	NotificationStatus bool   `json:"notificationStatus"`
}

type UserSettings struct {
	NotificationStatus bool `json:"notificationStatus"`
}

func (d UserDevice) UserDeviceToDBModel() *dbmodels.UserDevice {
	out := &dbmodels.UserDevice{
		Token:          d.Token,
		ImpartWealthID: d.ImpartWealthID,
		DeviceID:       d.DeviceID,
		DeviceToken:    d.DeviceToken,
		AppVersion:     d.AppVersion,
		DeviceName:     d.DeviceName,
		DeviceVersion:  d.DeviceVersion,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}

	return out
}

func UserDeviceFromDBModel(d *dbmodels.UserDevice) UserDevice {
	out := UserDevice{
		Token:          string(d.Token),
		ImpartWealthID: d.ImpartWealthID,
		DeviceID:       d.DeviceID,
		DeviceToken:    d.DeviceToken,
		AppVersion:     d.AppVersion,
		DeviceName:     d.DeviceName,
		DeviceVersion:  d.DeviceVersion,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
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
	Token          string
	DeviceID       string
	DeviceToken    string
	Negate         bool
}

// user Notification
type UserGlobalConfigInput struct {
	RefToken       string `json:"refToken,omitempty"`
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

type DeleteUserInput struct {
	ImpartWealthID string `json:"impartWealthID,omitempty"`
	Feedback       string `json:"feedback,omitempty"`
}
