package profile

import (
	"context"
	"database/sql"

	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"go.uber.org/zap"
)

type Store interface {
	GetUser(ctx context.Context, impartWealthID string) (*dbmodels.User, error)
	GetUserFromAuthId(ctx context.Context, authenticationId string) (*dbmodels.User, error)
	GetUserFromEmail(ctx context.Context, email string) (*dbmodels.User, error)
	GetUserFromScreenName(ctx context.Context, screenName string) (*dbmodels.User, error)
	CreateUserProfile(ctx context.Context, user *dbmodels.User, profile *dbmodels.Profile) error
	GetProfile(ctx context.Context, impartWealthId string) (*dbmodels.Profile, error)
	UpdateProfile(ctx context.Context, user *dbmodels.User, profile *dbmodels.Profile) error
	DeleteProfile(ctx context.Context, impartWealthID string, hardDelete bool) error
	DeleteUserProfile(ctx context.Context, gpi models.DeleteUserInput, hardDelete bool) impart.Error
	GetQuestionnaire(ctx context.Context, name string, version *uint) (*dbmodels.Questionnaire, error)
	GetAllCurrentQuestionnaires(ctx context.Context) (dbmodels.QuestionnaireSlice, error)
	GetUserQuestionnaires(ctx context.Context, impartWealthId string, questionnaireName *string) (dbmodels.QuestionnaireSlice, error)
	SaveUserQuestionnaire(ctx context.Context, answer dbmodels.UserAnswerSlice) error
	UpdateUserDemographic(ctx context.Context, answerIds []uint, status bool) error
	UpdateHiveUserDemographic(ctx context.Context, answerIds []uint, status bool, hiveid uint64) error

	GetUserConfigurations(ctx context.Context, impartWealthID string) (*dbmodels.UserConfiguration, error)
	CreateUserConfigurations(ctx context.Context, conf *dbmodels.UserConfiguration) (*dbmodels.UserConfiguration, error)
	EditUserConfigurations(ctx context.Context, conf *dbmodels.UserConfiguration) (*dbmodels.UserConfiguration, error)
	DeleteExceptUserDevice(ctx context.Context, impartID string, deviceToken string, refToken string) error

	GetUserDevice(ctx context.Context, token string, impartWealthID string, deviceID string) (*dbmodels.UserDevice, error)
	CreateUserDevice(ctx context.Context, device *dbmodels.UserDevice) (*dbmodels.UserDevice, error)
	UpdateDevice(ctx context.Context, device *dbmodels.UserDevice) error
	UpdateDeviceToken(ctx context.Context, device *dbmodels.UserDevice, deviceToken string) error

	GetUserNotificationMappData(input models.MapArgumentInput) (*dbmodels.NotificationDeviceMapping, error)
	CreateUserNotificationMappData(ctx context.Context, data *dbmodels.NotificationDeviceMapping) (*dbmodels.NotificationDeviceMapping, error)
	DeleteUserNotificationMappData(input models.MapArgumentInput) error
	UpdateExistingNotificationMappData(input models.MapArgumentInput, notifyStatus bool) error

	BlockUser(ctx context.Context, user *dbmodels.User, status bool) error
	GetMakeUp(ctx context.Context) (interface{}, error)
	GetUsersDetails(ctx context.Context, gpi models.GetAdminInputs) ([]models.UserDetail, *models.NextPage, error)
	GetPostDetails(ctx context.Context, gpi models.GetAdminInputs) ([]models.PostDetail, *models.NextPage, error)
	EditUserDetails(ctx context.Context, gpi models.WaitListUserInput) (string, impart.Error)
	GetHiveDetails(ctx context.Context, gpi models.GetAdminInputs) ([]map[string]interface{}, *models.NextPage, error)
	GetFilterDetails(ctx context.Context) ([]byte, error)
	EditBulkUserDetails(ctx context.Context, gpi models.UserUpdate) *models.UserUpdate
	DeleteBulkUserDetails(ctx context.Context, gpi models.UserUpdate) *models.UserUpdate

	GetAllUsers(ctx context.Context) error
}

func NewMySQLStore(db *sql.DB, logger *zap.Logger) Store {
	s := newMysqlStore(db, logger)

	return s
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

//func (s SearchType) IndexName() string {
//	switch s {
//	case emailSearch:
//		return emailIndexName
//	case screenNameSearch:
//		return screenNameIndexName
//	default:
//		return ""
//	}
//}

func EmailSearchType() SearchType {
	return emailSearch
}

func ScreenNameSearchType() SearchType {
	return screenNameSearch
}
