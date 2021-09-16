package plaid

import (
	"context"
	"database/sql"

	"github.com/impartwealthapp/backend/pkg/impart"
	"go.uber.org/zap"
)

var _ Service = &plaidHandler{}

type Service interface {
	SavePlaidInstitutions(ctx context.Context) error
	SavePlaidInstitutionToken(ctx context.Context, userInstitution UserInstitutionToken) impart.Error
	GetPlaidInstitutions(ctx context.Context) (Institutions, error)
	GetPlaidUserInstitutions(ctx context.Context, impartWealthId string) (UserInstitutionTokens, error)
	GetPlaidUserInstitutionAccounts(ctx context.Context, impartWealthId string) (UserAccount, impart.Error)
}

type plaidHandler struct {
	logger *zap.Logger
	db     *sql.DB
}

func NewPlaidService(db *sql.DB, logger *zap.Logger) Service {
	return &plaidHandler{
		logger: logger,
		db:     db,
	}
}
