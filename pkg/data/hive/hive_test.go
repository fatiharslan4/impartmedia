// +build integration

package data

import (
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/data/migrater"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/suite"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/zap"
	"testing"
)

type HiveTestSuite struct {
	suite.Suite
	logger   *zap.Logger
	cfg      *config.Impart
	db       *sql.DB
	hiveData HiveService
}

func TestHiveTestSuite(t *testing.T) {
	suite.Run(t, new(HiveTestSuite))
}

const pathToMigrationsDir = "../../../schemas/migrations"

func (s *HiveTestSuite) SetupSuite() {
	var err error
	s.logger = zap.NewNop()
	s.cfg, err = config.GetImpart()
	s.cfg.MigrationsPath = pathToMigrationsDir
	s.Require().NoError(err)
	s.db, err = s.cfg.GetDBConnection()
	s.Require().NoError(err)
	s.hiveData = NewHiveService(s.db, s.logger)

	migrationDB, err := s.cfg.GetMigrationDBConnection()
	s.Require().NoError(err)
	err = migrater.RunMigrationsDown(migrationDB, s.cfg.MigrationsPath, s.logger, nil)
	if err != nil && err != migrate.ErrNoChange {
		s.FailNow("DB Down Error %v", err)
	}
}

func (s *HiveTestSuite) TearDownSuite() {
	s.db.Close()
}

func (s *HiveTestSuite) SetupTest() {
	migrationDB, err := s.cfg.GetMigrationDBConnection()
	s.Require().NoError(err)
	defer migrationDB.Close()
	err = migrater.RunMigrationsUp(migrationDB, s.cfg.MigrationsPath, s.logger, nil)
	s.Require().NoError(err)
}

func (s *HiveTestSuite) TearDownTest() {
	migrationDB, err := s.cfg.GetMigrationDBConnection()
	s.Require().NoError(err)
	defer migrationDB.Close()
	err = migrater.RunMigrationsDown(migrationDB, s.cfg.MigrationsPath, s.logger, nil)
	s.Require().NoError(err)
}

func (s *HiveTestSuite) contextWithImpartAdmin() context.Context {
	id := ksuid.New().String()
	user := &dbmodels.User{
		ImpartWealthID:   id,
		AuthenticationID: id,
		Email:            id,
		ScreenName:       id,
		CreatedTS:        impart.CurrentUTC(),
		UpdatedTS:        impart.CurrentUTC(),
		DeviceToken:      "d",
		AwsSNSAppArn:     "f",
		Admin:            true,
	}
	err := user.Insert(context.TODO(), s.db, boil.Infer())
	s.NoError(err)

	return context.WithValue(context.Background(), impart.UserRequestContextKey, user)
}

func (s *HiveTestSuite) bootstrapTestHive(ctx context.Context) uint64 {

	hive := &dbmodels.Hive{
		Name:         "test hive",
		Description:  "test hive",
		PinnedPostID: null.Uint64From(1),
		//TagComparisons:       null.JSON{},
		NotificationTopicArn: null.StringFrom("abc"),
		//HiveDistributions:    null.JSON{},
	}
	hive, err := s.hiveData.NewHive(ctx, hive)
	s.NoError(err)
	return hive.HiveID
}

func (s *HiveTestSuite) TestCreateHive() {
	ctx := s.contextWithImpartAdmin()
	hive, err := s.hiveData.NewHive(ctx, &dbmodels.Hive{
		Name:         "test hive",
		Description:  "test hive",
		PinnedPostID: null.Uint64From(1),
		//TagComparisons:       null.JSON{},
		NotificationTopicArn: null.StringFrom("abc"),
		//HiveDistributions:    null.JSON{},
	})
	s.NoError(err)
	s.NotZero(hive.HiveID)
	gHive, err := s.hiveData.GetHive(ctx, hive.HiveID)
	s.NoError(err)
	s.Equal(gHive.HiveID, hive.HiveID)
}

func (s *HiveTestSuite) TestEditHive() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	existingHive, err := s.hiveData.GetHive(ctx, hiveID)
	s.NoError(err)
	s.Equal(existingHive.HiveID, hiveID)

	newPinnedPostID := null.Uint64From(existingHive.PinnedPostID.Uint64 + 1)
	existingHive.PinnedPostID = newPinnedPostID
	updatedHive, err := s.hiveData.EditHive(ctx, existingHive)

	s.NoError(err)
	s.Equal(newPinnedPostID.Uint64, updatedHive.PinnedPostID.Uint64)

}

func (s *HiveTestSuite) TestGetHives() {
	ctx := s.contextWithImpartAdmin()
	hiveID := s.bootstrapTestHive(ctx)
	hiveID2 := s.bootstrapTestHive(ctx)
	hiveID3 := s.bootstrapTestHive(ctx)

	hives, err := s.hiveData.GetHives(ctx)
	s.NoError(err)

	var foundHive1, foundHive2, foundHive3 bool
	for _, h := range hives {
		switch h.HiveID {
		case 1:

		case hiveID:
			foundHive1 = true
		case hiveID2:
			foundHive2 = true
		case hiveID3:
			foundHive3 = true
		default:
			s.FailNow("found an unexpected hive id %s", h.HiveID)
		}
	}
	s.True(foundHive1)
	s.True(foundHive2)
	s.True(foundHive3)
}
