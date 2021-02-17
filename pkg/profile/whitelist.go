package profile

import (
	"fmt"
	"strings"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"

	profiledata "github.com/impartwealthapp/backend/pkg/data/profile"
	"github.com/leebenson/conform"
	"go.uber.org/zap"
)

func (ps *profileService) CreateWhitelistEntry(whitelistEntry models.WhiteListProfile) impart.Error {
	if err := conform.Strings(&whitelistEntry); err != nil {
		return impart.NewError(err, "invalid message format")
	}
	err := ps.db.CreateWhitelistEntry(whitelistEntry)
	if err != nil {
		return impart.NewError(err, "unable to create whitelist entry")
	}
	return nil
}

func (ps *profileService) WhiteListSearch(impartWealthId, email, screenName string) (models.WhiteListProfile, impart.Error) {
	var out models.WhiteListProfile
	var err error
	email = strings.ToLower(email)
	screenName = strings.ToLower(screenName)

	if strings.TrimSpace(impartWealthId) != "" {
		ps.Logger().Debug("searching for", zap.String("impartWealthId", impartWealthId))
		out, err = ps.db.GetWhitelistEntry(impartWealthId)
	} else if strings.TrimSpace(email) != "" {
		ps.Logger().Debug("searching for", zap.String("email", email))
		out, err = ps.db.SearchWhitelistEntry(profiledata.EmailSearchType(), email)
	} else { // screenname search
		ps.Logger().Debug("searching for", zap.String("screenName", screenName))
		out, err = ps.db.SearchWhitelistEntry(profiledata.ScreenNameSearchType(), screenName)
	}

	if err != nil {
		if err == impart.ErrNotFound {
			return out, impart.NewError(err, "whitelist entry not found")
		}
		ps.SugaredLogger.Desugar().Error("error retrieving whitelist", zap.Error(err),
			zap.Strings("search params", []string{impartWealthId, screenName, email}))
		return out, impart.NewError(err, "error retrieving whitelist")
	}

	return out, nil
}

func (ps *profileService) UpdateWhitelistScreenName(impartWealthId, screenName string) impart.Error {
	screenName = strings.TrimSpace(strings.ToLower(screenName))
	existingWhitelist, err := ps.db.GetWhitelistEntry(impartWealthId)
	if err != nil {
		return impart.NewError(err, "unable to get existing whitelist entry for impartWealthId"+impartWealthId)
	}
	//noop
	if strings.TrimSpace(existingWhitelist.ScreenName) == strings.TrimSpace(screenName) {
		return nil
	}

	//validate the screen name is not already used
	existingImaprtWealthID, err := ps.db.GetImpartIdFromScreenName(screenName)
	if existingImaprtWealthID != "" || err != nil && err != impart.ErrNotFound {
		return impart.NewError(impart.ErrExists, fmt.Sprintf("Screen Name '%s' is already taken", screenName))
	}

	reservedWhitelistEntry, err := ps.db.SearchWhitelistEntry(profiledata.ScreenNameSearchType(), screenName)
	if (reservedWhitelistEntry.ImpartWealthID != "" && reservedWhitelistEntry.ImpartWealthID != impartWealthId) ||
		(err != nil && err != impart.ErrNotFound) {
		return impart.NewError(impart.ErrExists, fmt.Sprintf("Screen Name '%s' is already reserved", screenName))
	}

	err = ps.db.UpdateWhitelistEntryScreenName(impartWealthId, screenName)
	if err != nil {
		return impart.NewError(err, "error reserving ScreenName "+screenName)
	}

	return nil
}
