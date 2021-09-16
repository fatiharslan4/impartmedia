package plaid

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
)

var _ Service = &plaidHandler{}

type Service interface {
	SavePlaidInstitutions(ctx context.Context) error
	SavePlaidInstitutionToken(ctx context.Context, userInstitution UserInstitution) error
	GetPlaidInstitutions(ctx context.Context) (Institutions, error)
	GetPlaidUserInstitutions(ctx context.Context, impartWealthId string) (UserInstitutions, error)
}

type plaidHandler struct {
	logger *zap.Logger
	// hiveData            data.Hives
	// postData            data.Posts
	// commentData         data.Comments
	// reactionData        data.UserTrack
	// profileData         profiledata.Store
	// notificationService impart.NotificationService
	db *sql.DB
}

func NewPlaidService(db *sql.DB, logger *zap.Logger) Service {
	return &plaidHandler{
		logger: logger,
		db:     db,
	}
}
