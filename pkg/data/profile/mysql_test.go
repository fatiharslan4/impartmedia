// +build integration

package profile

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/data/migrater"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
	"time"
)

type ProfileDBTestSuite struct {
	suite.Suite
	logger *zap.Logger
	cfg    *config.Impart
	db     *sql.DB
	store  *mysqlStore
}

func TestProfileDBTestSuite(t *testing.T) {
	suite.Run(t, new(ProfileDBTestSuite))
}

const pathToMigrationsDir = "../../../schemas/migrations"

func (s *ProfileDBTestSuite) SetupSuite() {
	var err error
	s.logger = zap.NewNop()
	s.cfg, err = config.GetImpart()
	s.cfg.MigrationsPath = pathToMigrationsDir
	s.Require().NoError(err)
	s.db, err = s.cfg.GetDBConnection()
	s.Require().NoError(err)
	s.store = newMysqlStore(s.db, s.logger)

	migrationDB, err := s.cfg.GetMigrationDBConnection()
	s.Require().NoError(err)
	err = migrater.RunMigrationsDown(migrationDB, s.cfg.MigrationsPath, s.logger, nil)
	if err != nil && err != migrate.ErrNoChange {
		s.FailNow("DB Down Error %v", err)
	}
}

func (s *ProfileDBTestSuite) TearDownSuite() {
	s.db.Close()
}

func (s *ProfileDBTestSuite) SetupTest() {
	migrationDB, err := s.cfg.GetMigrationDBConnection()
	s.Require().NoError(err)
	defer migrationDB.Close()
	err = migrater.RunMigrationsUp(migrationDB, s.cfg.MigrationsPath, s.logger, nil)
	s.Require().NoError(err)
}

func (s *ProfileDBTestSuite) TearDownTest() {
	migrationDB, err := s.cfg.GetMigrationDBConnection()
	s.Require().NoError(err)
	defer migrationDB.Close()
	err = migrater.RunMigrationsDown(migrationDB, s.cfg.MigrationsPath, s.logger, nil)
	s.Require().NoError(err)
}

func (s *ProfileDBTestSuite) TestCreateProfile() {
	ctx := context.Background()
	id := ksuid.New().String()
	u := &dbmodels.User{
		ImpartWealthID:   id,
		AuthenticationID: fmt.Sprintf("auth0|%s", ksuid.New().String()),
		Email:            "test@test.com",
		ScreenName:       "testScreenName",
		CreatedTS:        time.Now().Truncate(time.Millisecond).UTC(),
		UpdatedTS:        time.Now().Truncate(time.Millisecond).UTC(),
		DeviceToken:      "device",
		AwsSNSAppArn:     "somearn",
	}
	p := &dbmodels.Profile{
		ImpartWealthID:  id,
		UpdatedTS:       time.Now().Truncate(time.Millisecond).UTC(),
		Attributes:      nil,
		SurveyResponses: nil,
	}

	attr := &models.Attributes{
		Name: "Philis",
	}
	err := p.Attributes.Marshal(attr)
	s.Require().NoError(err)

	survResp := &models.SurveyResponses{
		BirthYear: 2000,
	}
	err = p.SurveyResponses.Marshal(survResp)
	s.Require().NoError(err)

	err = s.store.CreateUserProfile(ctx, u, p)
	s.Require().NoError(err, "updated ts %v", u.UpdatedTS)

	retU, err := s.store.GetUser(ctx, u.ImpartWealthID)
	s.Require().NoError(err)
	s.Equal(u.ImpartWealthID, retU.ImpartWealthID)
	s.Equal(u.AuthenticationID, retU.AuthenticationID)
	s.Equal(u.Email, retU.Email)
	s.Equal(u.ScreenName, retU.ScreenName)
	s.Equal(u.CreatedTS, retU.CreatedTS)
	s.Equal(u.UpdatedTS, retU.UpdatedTS)
	s.Equal(u.DeviceToken, retU.DeviceToken)
	s.Equal(u.AwsSNSAppArn, retU.AwsSNSAppArn)

	s.Equal(p.ImpartWealthID, retU.R.ImpartWealthProfile.ImpartWealthID)

	var retAttributes models.Attributes
	err = retU.R.ImpartWealthProfile.Attributes.Unmarshal(&retAttributes)
	s.Require().NoError(err)

	var retSurvey models.SurveyResponses
	err = retU.R.ImpartWealthProfile.SurveyResponses.Unmarshal(&retSurvey)
	s.Require().NoError(err)

	s.Equal(*attr, retAttributes)
	s.Equal(*survResp, retSurvey)

}

func (s *ProfileDBTestSuite) TestFetches() {
	ctx := context.Background()
	id := ksuid.New().String()
	u := &dbmodels.User{
		ImpartWealthID:   id,
		AuthenticationID: fmt.Sprintf("auth0|%s", id),
		Email:            "test@test.com",
		ScreenName:       "testScreenName",
		CreatedTS:        time.Now().Truncate(time.Millisecond).UTC(),
		UpdatedTS:        time.Now().Truncate(time.Millisecond).UTC(),
	}
	p := &dbmodels.Profile{
		ImpartWealthID:  id,
		UpdatedTS:       time.Now().Truncate(time.Millisecond).UTC(),
		Attributes:      nil,
		SurveyResponses: nil,
	}

	err := p.Attributes.Marshal(&models.Attributes{
		Name: "Philis",
	})
	s.Require().NoError(err)

	err = p.SurveyResponses.Marshal(&models.SurveyResponses{
		BirthYear: 2000,
	})
	s.Require().NoError(err)

	err = s.store.CreateUserProfile(ctx, u, p)
	s.Require().NoError(err)

	retUser, err := s.store.GetUserFromAuthId(ctx, u.AuthenticationID)
	s.Equal(u.ImpartWealthID, retUser.ImpartWealthID)

	retUser, err = s.store.GetUserFromEmail(ctx, u.Email)
	s.Equal(u.ImpartWealthID, retUser.ImpartWealthID)

	retUser, err = s.store.GetUserFromScreenName(ctx, u.ScreenName)
	s.Equal(u.ImpartWealthID, retUser.ImpartWealthID)

	authUser, err := s.store.GetUserFromAuthId(ctx, u.AuthenticationID)
	s.Require().NoError(err)

	s.Equal(u.ImpartWealthID, authUser.ImpartWealthID)
	s.Equal(u.UpdatedTS, authUser.UpdatedTS)

	s.Equal(u.ImpartWealthID, authUser.R.ImpartWealthProfile.ImpartWealthID)

}

func (s *ProfileDBTestSuite) TestUpdate() {
	ctx := context.Background()
	id := ksuid.New().String()
	u := &dbmodels.User{
		ImpartWealthID:   id,
		AuthenticationID: fmt.Sprintf("auth0|%s", ksuid.New().String()),
		Email:            "test@test.com",
		ScreenName:       "testScreenName",
		CreatedTS:        time.Now().Truncate(time.Millisecond).UTC(),
		UpdatedTS:        time.Now().Truncate(time.Millisecond).UTC(),
	}

	p := &dbmodels.Profile{
		ImpartWealthID:  id,
		UpdatedTS:       time.Now().Truncate(time.Millisecond).UTC(),
		Attributes:      nil,
		SurveyResponses: nil,
	}
	attr := &models.Attributes{
		Name: "Philis",
	}
	err := p.Attributes.Marshal(attr)
	s.Require().NoError(err)

	survResp := &models.SurveyResponses{
		BirthYear: 2000,
	}
	err = p.SurveyResponses.Marshal(survResp)
	s.Require().NoError(err)

	err = s.store.CreateUserProfile(ctx, u, p)
	s.Require().NoError(err)

	u.ScreenName = "differentScreenName"
	err = s.store.UpdateProfile(ctx, u, nil)
	s.Require().NoError(err)
	retU, err := s.store.GetUser(ctx, u.ImpartWealthID)
	s.Equal("differentScreenName", retU.ScreenName)

	attr = &models.Attributes{
		Name: "newName",
	}
	err = p.Attributes.Marshal(attr)
	s.Require().NoError(err)
	err = s.store.UpdateProfile(ctx, nil, p)

	newAttr := &models.Attributes{}
	retP, err := s.store.GetUser(ctx, u.ImpartWealthID)
	s.Require().NoError(err)

	err = retP.R.ImpartWealthProfile.Attributes.Unmarshal(&newAttr)
	s.Require().NoError(err)
	s.Equal(attr.Name, newAttr.Name)

}

func (s *ProfileDBTestSuite) TestDelete() {
	ctx := context.Background()
	id := ksuid.New().String()
	u := &dbmodels.User{
		ImpartWealthID:   id,
		AuthenticationID: fmt.Sprintf("auth0|%s", ksuid.New().String()),
		Email:            "test@test.com",
		ScreenName:       "testScreenName",
		CreatedTS:        time.Now().Truncate(time.Millisecond).UTC(),
		UpdatedTS:        time.Now().Truncate(time.Millisecond).UTC(),
	}
	p := &dbmodels.Profile{
		ImpartWealthID:  id,
		UpdatedTS:       time.Now().Truncate(time.Millisecond).UTC(),
		Attributes:      nil,
		SurveyResponses: nil,
	}

	attr := &models.Attributes{
		Name: "Philis",
	}
	err := p.Attributes.Marshal(attr)
	s.Require().NoError(err)

	survResp := &models.SurveyResponses{
		BirthYear: 2000,
	}
	err = p.SurveyResponses.Marshal(survResp)
	s.Require().NoError(err)

	err = s.store.CreateUserProfile(ctx, u, p)
	s.Require().NoError(err)

	err = s.store.DeleteProfile(ctx, u.ImpartWealthID)
	s.Require().NoError(err)

	err = s.store.DeleteProfile(ctx, u.ImpartWealthID)
	s.Require().Equal(impart.ErrNotFound, err)

	_, err = s.store.GetUser(ctx, u.ImpartWealthID)
	s.Require().Equal(impart.ErrNotFound, err)
}
