package plaid

import (
	"context"
	"database/sql"

	hive "github.com/impartwealthapp/backend/pkg/hive"
	"github.com/impartwealthapp/backend/pkg/impart"
	"go.uber.org/zap"
)

var _ Service = &service{}

type Service interface {
	SavePlaidInstitutions(ctx context.Context) error
	SavePlaidInstitutionToken(ctx context.Context, userInstitution UserInstitutionToken) impart.Error
	GetPlaidInstitutions(ctx context.Context) (Institutions, error)
	GetPlaidUserInstitutions(ctx context.Context, impartWealthId string) (UserInstitutionTokens, error)
	GetPlaidUserInstitutionAccounts(ctx context.Context, impartWealthId string) (UserAccount, impart.Error)
}

type service struct {
	logger *zap.Logger
	Hive   hive.Service
	db     *sql.DB
}

// //  Create New Plaid Service

func New(db *sql.DB, logger *zap.Logger, hive hive.Service) Service {
	svc := &service{
		logger: logger,
		db:     db,
		Hive:   hive,
	}

	return svc

}
