package profile

import (
	"context"
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"github.com/impartwealthapp/backend/pkg/data/types"
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

func (ps *profileService) ValidateSchema(document gojsonschema.JSONLoader) (errors []impart.Error) {
	result, err := gojsonschema.Validate(ps.schemaValidator, document)
	if err != nil {
		ps.SugaredLogger.Error(err.Error())
		return []impart.Error{
			impart.NewError(impart.ErrBadRequest, "unable to validate schema"),
		}
	}

	if result.Valid() {
		return nil
	}
	// msg := fmt.Sprintf("%v validations errors.\n", len(result.Errors()))
	msg := "validations errors"
	for i, desc := range result.Errors() {
		msg += fmt.Sprintf("%v: %s\n", i, desc)
		er := impart.NewError(impart.ErrValidationError, fmt.Sprintf("%s ", desc), impart.ErrorKey(desc.Field()))
		errors = append(errors, er)
	}
	return errors
}

func (ps *profileService) ValidateScreenNameInput(document gojsonschema.JSONLoader) (errors []impart.Error) {
	v := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", "./schemas/json/ScreenNameValidator.json"))
	_, err := v.LoadJSON()
	if err != nil {
		ps.SugaredLogger.Error(err.Error())
		return []impart.Error{
			impart.NewError(impart.ErrBadRequest, "unable to load validation schema"),
		}
	}
	result, err := gojsonschema.Validate(v, document)
	if err != nil {
		ps.SugaredLogger.Error(err.Error())
		return []impart.Error{
			impart.NewError(impart.ErrBadRequest, "unable to validate schema"),
		}
	}

	if result.Valid() {
		return nil
	}
	// msg := fmt.Sprintf("%v validations errors.\n", len(result.Errors()))
	msg := "validations errors"
	for i, desc := range result.Errors() {
		msg += fmt.Sprintf("%v: %s\n", i, desc)
		er := impart.NewError(impart.ErrValidationError, fmt.Sprintf("%s ", desc), impart.ErrorKey(desc.Field()))
		errors = append(errors, er)
	}
	return errors
}

func (ps *profileService) ValidateInput(document gojsonschema.JSONLoader, validationModel types.Type) (errors []impart.Error) {

	v := gojsonschema.NewReferenceLoader(
		fmt.Sprintf("file://%s", "./schemas/json/"+validationModel+".json"),
	)
	_, err := v.LoadJSON()
	if err != nil {
		ps.SugaredLogger.Error(err.Error())
		return []impart.Error{
			impart.NewError(impart.ErrBadRequest, "unable to load validation schema"),
		}
	}
	result, err := gojsonschema.Validate(v, document)
	if err != nil {
		ps.SugaredLogger.Error(err.Error())
		return []impart.Error{
			impart.NewError(impart.ErrBadRequest, "unable to validate schema"),
		}
	}

	if result.Valid() {
		return nil
	}
	// msg := fmt.Sprintf("%v validations errors.\n", len(result.Errors()))
	msg := "validations errors"
	for i, desc := range result.Errors() {
		msg += fmt.Sprintf("%v: %s\n", i, desc)
		er := impart.NewError(impart.ErrValidationError, fmt.Sprintf("%s ", desc), impart.ErrorKey(desc.Field()))
		errors = append(errors, er)
	}
	return errors
}

/**
 *
 * Validate the screen name
 *	No screen names can contain Impart, Impartwealth,
 *  mod, moderator or Admin unless they are official Impart Wealth account.
 *
 */
func (ps *profileService) ValidateScreenNameString(ctx context.Context, screenName string) impart.Error {
	var invalidStrings = []string{
		"impart", "impartwealth", "mod", "moderator", "admin", "wealth",
	}
	var err impart.Error
	for _, str := range invalidStrings {
		if ok := strings.Index(strings.ToLower(screenName), str); ok > -1 {
			err = impart.NewError(impart.ErrValidationError, "this screen name is not allowed.")
			break
		}
	}

	return err
}
