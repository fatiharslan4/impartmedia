package models

import (
	"context"
	"fmt"
	"time"

	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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
	DeleteByAdmin  bool   `json:"deleteByAdmin,omitempty"`
}

type WaitListUserInput struct {
	ImpartWealthID string `json:"impartWealthID,omitempty"`
	Type           string `json:"type,omitempty"`
	HiveID         uint64 `json:"hiveID,omitempty"`
}

func UpdateToUserDB(userToDelete *dbmodels.User, gpi DeleteUserInput, isDelete bool, screenName string, userEmail string) *dbmodels.User {
	if isDelete {
		userToDelete.Feedback = null.StringFromPtr(&gpi.Feedback)
		currTime := time.Now().In(boil.GetLocation())
		userToDelete.DeletedAt = null.TimeFrom(currTime)
		userToDelete.ScreenName = fmt.Sprintf("%s-%s", userToDelete.ScreenName, userToDelete.ImpartWealthID)
		userToDelete.Email = fmt.Sprintf("%s-%s", userToDelete.Email, userToDelete.ImpartWealthID)
		userToDelete.DeletedByAdmin = gpi.DeleteByAdmin
	} else {
		userToDelete.ScreenName = screenName
		userToDelete.Feedback = null.String{}
		userToDelete.DeletedAt = null.Time{}
		userToDelete.Email = userEmail
		userToDelete.DeletedByAdmin = false
	}
	return userToDelete

}

type PlaidInput struct {
	ImpartWealthID   string `json:"impartWealthID"`
	PlaidAccessToken string `json:"plaidAccessToken"`
}
