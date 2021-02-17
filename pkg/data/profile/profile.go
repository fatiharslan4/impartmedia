package profile

import "github.com/impartwealthapp/backend/pkg/models"

type Store interface {
	GetImpartIdFromAuthId(authenticationId string) (string, error)
	GetImpartIdFromEmail(email string) (string, error)
	GetImpartIdFromScreenName(screenName string) (string, error)
	CreateProfile(p models.Profile) (models.Profile, error)
	GetProfile(impartWealthId string, consistentRead bool) (models.Profile, error)
	GetProfileFromAuthId(authenticationId string, consistentRead bool) (models.Profile, error)
	UpdateProfile(authId string, p models.Profile) (models.Profile, error)
	DeleteProfile(impartWealthID string) error

	GetNotificationProfiles(*models.NextProfilePage) ([]models.Profile, *models.NextProfilePage, error)

	UpdateProfileProperty(impartWealthID, propertyPathName string, propertyPathValue interface{}) error

	CreateWhitelistEntry(profile models.WhiteListProfile) error
	GetWhitelistEntry(impartWealthID string) (models.WhiteListProfile, error)
	SearchWhitelistEntry(t SearchType, value string) (models.WhiteListProfile, error)
	UpdateWhitelistEntryScreenName(impartWealthID, screenName string) error

	GetHive(hiveID string, consistentRead bool) (models.Hive, error)
}

// The type of search to perform
type SearchType int

const (
	emailSearch SearchType = iota
	screenNameSearch
)

func (s SearchType) String() string {
	switch s {
	case emailSearch:
		return "email"
	case screenNameSearch:
		return "screenName"
	default:
		return ""
	}
}

func (s SearchType) IndexName() string {
	switch s {
	case emailSearch:
		return emailIndexName
	case screenNameSearch:
		return screenNameIndexName
	default:
		return ""
	}
}

func EmailSearchType() SearchType {
	return emailSearch
}

func ScreenNameSearchType() SearchType {
	return screenNameSearch
}
