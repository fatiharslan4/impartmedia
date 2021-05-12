package profile

import (
	"context"
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/segmentio/ksuid"
	"github.com/xeipuuv/gojsonschema"
)

var screenNameRegexp = regexp.MustCompile(`[[:alnum:]]+$`)

func (ps *profileService) validateNewProfile(ctx context.Context, p models.Profile) impart.Error {
	var err error

	if _, err = ksuid.Parse(p.ImpartWealthID); err != nil {
		return impart.NewError(err, "invalid impartWealthId format")
	}

	// Validate doesn't exist
	user, err := ps.profileStore.GetUser(ctx, p.ImpartWealthID)
	if err != nil {
		if err != impart.ErrNotFound {
			return impart.NewError(err, "error checking impartID")
		}
		err = nil
	}
	if user != nil {
		return impart.NewError(impart.ErrExists, "impartWealthId already exists")
	}

	// Validate Auth ID
	if strings.TrimSpace(p.AuthenticationID) == "" {
		return impart.NewError(impart.ErrBadRequest, "authenticationId is required")
	}

	_, err = mail.ParseAddress(p.Email)
	if err != nil {
		return impart.NewError(impart.ErrBadRequest, "invalid email address", impart.Email)
	}

	user, err = ps.profileStore.GetUserFromAuthId(ctx, p.AuthenticationID)
	if err != nil {
		if err != impart.ErrNotFound {
			return impart.NewError(err, "error checking authenticationId")
		} else {
			err = nil
		}
	}
	if user != nil {
		return impart.NewError(impart.ErrExists, fmt.Sprintf("authenticationId %s already exists!", p.AuthenticationID))
	}

	user, err = ps.profileStore.GetUserFromEmail(ctx, p.Email)
	if err != nil {
		if err != impart.ErrNotFound {
			return impart.NewError(err, "error checking email", impart.Email)
		} else {
			err = nil
		}
	}
	if user != nil {
		return impart.NewError(impart.ErrExists, "email already exists!", impart.Email)
	}

	if screenNameRegexp.FindString(p.ScreenName) != p.ScreenName {
		{
			ps.Logger().Error("invalid screen name", zap.String("screenName", p.ScreenName))
			return impart.NewError(impart.ErrBadRequest, "invalid screen name, must be alphanumeric characters only", impart.ScreenName)
		}
	}
	user, err = ps.profileStore.GetUserFromScreenName(ctx, p.ScreenName)
	if err != nil {
		if err != impart.ErrNotFound {
			return impart.NewError(err, "error checking screenName", impart.ScreenName)
		} else {
			err = nil
		}
	}
	if user != nil {
		return impart.NewError(impart.ErrExists, "screenName already exists", impart.ScreenName)
	}

	return nil
}

func (ps *profileService) ValidateSchema(document gojsonschema.JSONLoader) impart.Error {
	result, err := gojsonschema.Validate(ps.schemaValidator, document)
	if err != nil {
		ps.SugaredLogger.Error(err.Error())
		return impart.NewError(impart.ErrBadRequest, "unable to validate schema")
	}

	if result.Valid() {
		return nil
	}
	msg := fmt.Sprintf("%v validations errors.\n", len(result.Errors()))
	for i, desc := range result.Errors() {
		msg += fmt.Sprintf("%v: %s\n", i, desc)
	}

	return impart.NewError(impart.ErrBadRequest, msg)
}
