package plaid

import (
	"context"
	"database/sql"

	hive "github.com/impartwealthapp/backend/pkg/hive"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"go.uber.org/zap"
)

var _ Service = &service{}

type Service interface {
	SavePlaidInstitutions(ctx context.Context) error
	SavePlaidInstitutionToken(ctx context.Context, userInstitution UserInstitutionToken) impart.Error
	GetPlaidInstitutions(ctx context.Context) (Institutions, error)
	GetPlaidUserInstitutions(ctx context.Context, impartWealthId string) (UserInstitutionTokens, error)
	GetPlaidUserInstitutionAccounts(ctx context.Context, impartWealthId string, gpi models.GetPlaidInput) (UserAccount, *NextPage, impart.Error)
	GetPlaidUserInstitutionTransactions(ctx context.Context, impartWealthId string, gpi models.GetPlaidInput) (UserTransaction, *NextPage, []PlaidError)
	GetPlaidUserAccountsTransactions(ctx context.Context, accountId string, userInstitutionId uint64, impartWealthId string, gpi models.GetPlaidAccountTransactionInput) (UserTransaction, *NextPage, []PlaidError)
	DeletePlaidUserInstitutionAccounts(ctx context.Context, userInstitutionId uint64) impart.Error
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
