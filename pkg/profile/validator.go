package profile

import (
	"context"
	"encoding/json"
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

func (ps *profileService) validateNewProfile(ctx context.Context, p models.Profile, apiVersion string) impart.Error {
	var err error

	if _, err = ksuid.Parse(p.ImpartWealthID); err != nil {
		return impart.NewError(err, "invalid impartWealthId format")
	}

	if apiVersion == "v1.1" {
		if len(strings.TrimSpace(p.FirstName)) == 0 {
			return impart.NewError(impart.ErrBadRequest, string(impart.FirstNameRequired))
		}
		if len(strings.TrimSpace(p.LastName)) == 0 {
			return impart.NewError(impart.ErrBadRequest, string(impart.LastNameRequired))
		}
	} else if apiVersion == "v1" {
		if len(strings.TrimSpace(p.Attributes.Name)) == 0 {
			return impart.NewError(impart.ErrBadRequest, string(impart.NameRequired))
		}
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
		return impart.NewError(impart.ErrBadRequest, "Email already in use.", impart.Email)
	}

	if screenNameRegexp.FindString(p.ScreenName) != p.ScreenName {
		{
			ps.Logger().Error("invalid screen name", zap.String("screenName", p.ScreenName))
			return impart.NewError(impart.ErrBadRequest, "Invalid characters, please use letters and numbers only.", impart.ScreenName)
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

func (ps *profileService) ValidateScreenNameInput(document gojsonschema.JSONLoader, screenName []byte) (errors []impart.Error) {
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
	screenname := models.ScreenNameValidator{}
	err = json.Unmarshal(screenName, &screenname)

	msg := "validations errors"
	msgDes := ""
	for i, desc := range result.Errors() {
		if len(strings.TrimSpace(screenname.ScreenName)) < 8 {
			msgDes += "Screen name too short, must be 8 or more characters"
		} else if len(strings.TrimSpace(screenname.ScreenName)) > 15 {
			msgDes += "Screen name too long, must be 15 or less characters"
		} else {
			msg += fmt.Sprintf("%v: %s\n", i, desc)
			msgDes = fmt.Sprintf("%s ", desc)
		}
		er := impart.NewError(impart.ErrValidationError, msgDes, impart.ErrorKey(desc.Field()))
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

// Validate the screen name
// No screen names can contain Impart, Impartwealth,
// mod, moderator or Admin unless they are official Impart Wealth account.
func (ps *profileService) ValidateScreenNameString(ctx context.Context, screenName string) impart.Error {
	var invalidStrings = []string{
		"impart", "impartwealth", "mod", "moderator", "admin", "wealth",
	}
	var err impart.Error
	for _, str := range invalidStrings {
		if ok := strings.Index(strings.ToLower(screenName), str); ok > -1 {
			err = impart.NewError(impart.ErrValidationError, "Screen name includes invalid terms.")
			break
		}
	}

	return err
}
