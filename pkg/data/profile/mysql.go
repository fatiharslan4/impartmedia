package profile

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

var _ Store = &mysqlStore{}

type mysqlStore struct {
	logger *zap.Logger
	db     *sql.DB
}

func (m mysqlStore) GetProfile(ctx context.Context, impartWealthId string) (*dbmodels.Profile, error) {
	return dbmodels.Profiles(Where("impart_wealth_id = ?", impartWealthId)).One(ctx, m.db)
}

func newMysqlStore(db *sql.DB, logger *zap.Logger) *mysqlStore {
	out := &mysqlStore{
		db:     db,
		logger: logger,
	}
	return out
}

func (m mysqlStore) getUser(ctx context.Context, impartID, authID, email, screenName string) (*dbmodels.User, error) {
	var clause QueryMod
	if impartID != "" {
		clause = Where(fmt.Sprintf("%s = ?", dbmodels.UserColumns.ImpartWealthID), impartID)
	} else if authID != "" {
		clause = Where(fmt.Sprintf("%s = ?", dbmodels.UserColumns.AuthenticationID), authID)
	} else if email != "" {
		clause = Where(fmt.Sprintf("%s = ?", dbmodels.UserColumns.Email), email)
	} else {
		clause = Where(fmt.Sprintf("%s = ?", dbmodels.UserColumns.ScreenName), screenName)
	}
	usersWhere := []QueryMod{
		clause,
		Load(dbmodels.UserRels.ImpartWealthProfile),
		Load(dbmodels.UserRels.MemberHiveHives),
	}

	u, err := dbmodels.Users(usersWhere...).One(ctx, m.db)
	if err == sql.ErrNoRows {
		return nil, impart.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, err
}

func (m mysqlStore) GetUser(ctx context.Context, impartWealthID string) (*dbmodels.User, error) {
	return m.getUser(ctx, impartWealthID, "", "", "")
}

func (m mysqlStore) GetUserFromAuthId(ctx context.Context, authenticationId string) (*dbmodels.User, error) {
	return m.getUser(ctx, "", authenticationId, "", "")
}

func (m mysqlStore) GetUserFromEmail(ctx context.Context, email string) (*dbmodels.User, error) {
	return m.getUser(ctx, "", "", email, "")
}

func (m mysqlStore) GetUserFromScreenName(ctx context.Context, screenName string) (*dbmodels.User, error) {
	return m.getUser(ctx, "", "", "", screenName)
}

func rollbackIfError(tx *sql.Tx, err error, logger *zap.Logger) error {
	rErr := tx.Rollback()
	if rErr != nil {
		logger.Error("unable to rollback transaction", zap.Error(rErr))
		return fmt.Errorf(rErr.Error(), err)
	}
	return err
}

func (m mysqlStore) CreateUserProfile(ctx context.Context, user *dbmodels.User, profile *dbmodels.Profile) error {
	if user == nil || profile == nil {
		m.logger.Error("user or profile is nil")
		return impart.ErrBadRequest
	}
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return rollbackIfError(tx, err, m.logger)
	}
	if err = user.Insert(ctx, m.db, boil.Infer()); err != nil {
		return rollbackIfError(tx, err, m.logger)
	}
	if err = user.SetImpartWealthProfile(ctx, m.db, true, profile); err != nil {
		return rollbackIfError(tx, err, m.logger)
	}
	//add the default hive
	if err = user.SetMemberHiveHives(ctx, m.db, false, &dbmodels.Hive{HiveID: 1}); err != nil {
		return rollbackIfError(tx, err, m.logger)
	}
	return tx.Commit()
}

func (m mysqlStore) UpdateProfile(ctx context.Context, user *dbmodels.User, profile *dbmodels.Profile) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return rollbackIfError(tx, err, m.logger)
	}
	if user != nil {
		_, err = user.Update(ctx, tx, boil.Infer())
		if err != nil {
			return rollbackIfError(tx, err, m.logger)
		}
	}
	if profile != nil {
		_, err = profile.Update(ctx, tx, boil.Infer())
		if err != nil {
			return rollbackIfError(tx, err, m.logger)
		}
	}
	return tx.Commit()
}

func (m mysqlStore) DeleteProfile(ctx context.Context, impartWealthID string) error {
	u, err := dbmodels.FindUser(ctx, m.db, impartWealthID)
	if err == sql.ErrNoRows || u == nil {
		return impart.ErrNotFound
	}
	if err != nil {
		return err
	}
	_, err = u.Delete(ctx, m.db)
	return err
}
