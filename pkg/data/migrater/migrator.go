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

func BootStrapAdminUsers(db *sql.DB, env config.Environment, logger *zap.Logger) error {
	bootstrapAdmins := envAdmins[env]
	hives, err := dbmodels.Hives().All(context.TODO(), db)
	if err != nil {
		return err
	}

	logger.Info("bootstrapping admin users", zap.String("environment", string(env)),
		zap.Int("hiveCount", len(hives)), zap.Int("userCount", len(bootstrapAdmins)))
	for _, u := range bootstrapAdmins {
		tx, err := db.BeginTx(context.TODO(), &sql.TxOptions{
			Isolation: sql.LevelSerializable,
		})
		if err != nil {
			if tx != nil {
				tx.Rollback()
			}
			return err
		}
		logger.Info("Bootstrapping user", zap.String("email", u.Email))
		if err = u.Upsert(context.TODO(), tx, boil.Infer(), boil.Infer()); err != nil {
			tx.Rollback()
			return err
		}
		//make sure the admin is a member of all hives

		if err = u.SetMemberHiveHives(context.TODO(), tx, false, hives...); err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
	}
	return nil
}

var envAdmins = map[config.Environment][]*dbmodels.User{
	config.Local: []*dbmodels.User{
		&dbmodels.User{
			ImpartWealthID:   "1pe1n5BsNo7COEkJXhZo7ubL0Fa",
			AuthenticationID: "auth0|604b0291d42b9200691ec8a4",
			Email:            "jamison@impart.media",
			ScreenName:       "1pe1n5BsNo7COEkJXhZo7ubL0Fa",
			CreatedTS:        impart.CurrentUTC(),
			UpdatedTS:        impart.CurrentUTC(),
			DeviceToken:      "",
			AwsSNSAppArn:     "",
			Admin:            true,
		},
	},
	config.Development: []*dbmodels.User{
		&dbmodels.User{
			ImpartWealthID:   "1pe1n5BsNo7COEkJXhZo7ubL0Fa",
			AuthenticationID: "auth0|604e8a730f2d99006dd9521d",
			Email:            "jamison@impart.media",
			ScreenName:       "1pe1n5BsNo7COEkJXhZo7ubL0Fa",
			CreatedTS:        impart.CurrentUTC(),
			UpdatedTS:        impart.CurrentUTC(),
			DeviceToken:      "",
			AwsSNSAppArn:     "",
			Admin:            true,
		},
	},
	config.IOS: []*dbmodels.User{
		&dbmodels.User{
			ImpartWealthID:   "1pe1n5BsNo7COEkJXhZo7ubL0Fa",
			AuthenticationID: "auth0|604e8b2646a2f7007123d28d",
			Email:            "jamison@impart.media",
			ScreenName:       "1pe1n5BsNo7COEkJXhZo7ubL0Fa",
			CreatedTS:        impart.CurrentUTC(),
			UpdatedTS:        impart.CurrentUTC(),
			DeviceToken:      "",
			AwsSNSAppArn:     "",
			Admin:            true,
		},
	},
	config.Preproduction: []*dbmodels.User{
		&dbmodels.User{
			ImpartWealthID:   "1pe1n5BsNo7COEkJXhZo7ubL0Fa",
			AuthenticationID: "auth0|604e8a87ffc20800689e61f2",
			Email:            "jamison@impart.media",
			ScreenName:       "1pe1n5BsNo7COEkJXhZo7ubL0Fa",
			CreatedTS:        impart.CurrentUTC(),
			UpdatedTS:        impart.CurrentUTC(),
			DeviceToken:      "",
			AwsSNSAppArn:     "",
			Admin:            true,
		},
	},
	config.Production: []*dbmodels.User{
		&dbmodels.User{
			ImpartWealthID:   "1pe1n5BsNo7COEkJXhZo7ubL0Fa",
			AuthenticationID: "auth0|604e8b0243674100696d7812",
			Email:            "jamison@impart.media",
			ScreenName:       "1pe1n5BsNo7COEkJXhZo7ubL0Fa",
			CreatedTS:        impart.CurrentUTC(),
			UpdatedTS:        impart.CurrentUTC(),
			DeviceToken:      "",
			AwsSNSAppArn:     "",
			Admin:            true,
		},
	},
}
