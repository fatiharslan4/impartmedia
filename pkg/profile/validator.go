package profile

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/segmentio/ksuid"
	"github.com/xeipuuv/gojsonschema"
)

func (ps *profileService) validateNewProfile(p models.Profile) impart.Error {
	var err error

	if _, err = ksuid.Parse(p.ImpartWealthID); err != nil {
		return impart.NewError(err, "invalid impartWealthId - must be ksuid")
	}

	// Validate doesn't exist
	existingProfile, err := ps.db.GetProfile(p.ImpartWealthID, true)
	if err != nil {
		if err != impart.ErrNotFound {
			return impart.NewError(err, "error checking impartID")
		}
		err = nil
	}
	if existingProfile.ImpartWealthID != "" {
		return impart.NewError(impart.ErrExists, "profile already exists")
	}

	// Validate Auth ID
	if strings.TrimSpace(p.AuthenticationID) == "" {
		return impart.NewError(impart.ErrBadRequest, "authenticationId is required")
	}

	_, err = mail.ParseAddress(p.Email)
	if err != nil {
		return impart.NewError(impart.ErrBadRequest, "invalid email address")
	}

	impartWealthId, err := ps.db.GetImpartIdFromAuthId(p.AuthenticationID)
	if err != nil {
		if err != impart.ErrNotFound {
			return impart.NewError(err, "error checking authenticationId")
		} else {
			err = nil
		}
	}
	if impartWealthId != "" {
		return impart.NewError(impart.ErrExists, fmt.Sprintf("authenticationId %s already exists!", p.AuthenticationID))
	}

	impartWealthId, err = ps.db.GetImpartIdFromEmail(p.Email)
	if err != nil {
		if err != impart.ErrNotFound {
			return impart.NewError(err, "error checking email")
		} else {
			err = nil
		}
	}
	if impartWealthId != "" {
		return impart.NewError(impart.ErrExists, "email already exists!")
	}

	if strings.TrimSpace(p.ScreenName) != "" {
		impartWealthId, err = ps.db.GetImpartIdFromScreenName(p.ScreenName)
		if err != nil {
			if err != impart.ErrNotFound {
				return impart.NewError(err, "error checking screenName")
			} else {
				err = nil
			}
		}
		if impartWealthId != "" {
			return impart.NewError(impart.ErrExists, "screenName already exists")
		}
	}

	return nil
}

func (ps *profileService) ValidateSchema(document gojsonschema.JSONLoader) impart.Error {
	result, err := gojsonschema.Validate(ps.schemaValidator, document)
	if err != nil {
		ps.SugaredLogger.Fatal(err.Error())
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
