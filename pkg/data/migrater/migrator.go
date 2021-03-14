package migrater

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/impartwealthapp/backend/internal/pkg/impart/config"
	"github.com/impartwealthapp/backend/pkg/impart"
	"github.com/impartwealthapp/backend/pkg/models/dbmodels"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const ImpartDBName = "impart"

var _ migrate.Logger = &migrateZapLogger{}

type migrateZapLogger struct {
	logger *zap.Logger
}

func (m *migrateZapLogger) Printf(format string, v ...interface{}) {
	m.logger.Info(fmt.Sprintf(format, v...))
}

func (m *migrateZapLogger) Verbose() bool {
	return m.logger.Core().Enabled(zapcore.DebugLevel)
}

func RunMigrationsUp(db *sql.DB, filepath string, logger *zap.Logger, stop <-chan bool) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{DatabaseName: ""})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", filepath),
		ImpartDBName,
		driver,
	)
	defer driver.Close()
	defer m.Close()
	if err != nil {
		return err
	}
	m.Log = &migrateZapLogger{logger: logger}
	if stop != nil {
		go func() {
			m.GracefulStop <- <-stop
		}()
	}
	err = m.Up()
	if err == migrate.ErrNoChange {
		return nil
	}
	return err
}

func RunMigrationsDown(db *sql.DB, filepath string, logger *zap.Logger, stop <-chan bool) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{DatabaseName: ""})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", filepath),
		ImpartDBName,
		driver,
	)
	defer driver.Close()
	defer m.Close()
	if err != nil {
		return err
	}
	m.Log = &migrateZapLogger{logger: logger}
	if stop != nil {
		go func() {
			m.GracefulStop <- <-stop
		}()
	}

	return m.Down()
}

func BootStrapAdminUser(db *sql.DB, env config.Environment, logger *zap.Logger) error {
	var bootstrapUser *dbmodels.User
	switch env {
	case config.Local:
		bootstrapUser = &dbmodels.User{
			ImpartWealthID:   "1pe1n5BsNo7COEkJXhZo7ubL0Fa",
			AuthenticationID: "auth0|604b0291d42b9200691ec8a4",
			Email:            "jamison@impart.media",
			ScreenName:       "1pe1n5BsNo7COEkJXhZo7ubL0Fa",
			CreatedTS:        impart.CurrentUTC(),
			UpdatedTS:        impart.CurrentUTC(),
			DeviceToken:      "",
			AwsSNSAppArn:     "",
			Admin:            true,
		}
	default:
		return nil
	}
	logger.Info("Bootstrapping user", zap.String("email", bootstrapUser.Email))
	err := bootstrapUser.Upsert(context.TODO(), db, boil.Infer(), boil.Infer())
	if err != nil {
		return err
	}
	bootstrapUser.SetMemberHiveHives(context.TODO(), db, false, &dbmodels.Hive{HiveID: 1})
	if err != nil {
		return err
	}
	return bootstrapUser.Upsert(context.TODO(), db, boil.Infer(), boil.Infer())
}
