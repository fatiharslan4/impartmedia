package config

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

func getDefaultConfig() *mysql.Config {
	mysqlConfig := mysql.NewConfig()
	mysqlConfig.Net = "tcp"
	mysqlConfig.AllowNativePasswords = true
	mysqlConfig.ParseTime = true
	mysqlConfig.CheckConnLiveness = true
	mysqlConfig.Collation = "utf8mb4_unicode_ci"
	mysqlConfig.Loc = time.UTC
	mysqlConfig.TLSConfig = "preferred"
	mysqlConfig.Timeout = 10 * time.Second
	mysqlConfig.ReadTimeout = 10 * time.Second
	mysqlConfig.WriteTimeout = 10 * time.Second
	mysqlConfig.Params = make(map[string]string)
	mysqlConfig.Params["charset"] = "utf8mb4"
	mysqlConfig.MultiStatements = true
	return mysqlConfig
}

func (ic *Impart) GetDBConnection() (*sql.DB, error) {

	mysqlConfig := getDefaultConfig()
	mysqlConfig.Addr = fmt.Sprintf("%s:%v", ic.DBHost, ic.DBPort)
	mysqlConfig.User = ic.DBUsername
	mysqlConfig.Passwd = ic.DBPassword
	mysqlConfig.DBName = ic.DBName

	db, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		return nil, err
	}
	if ic.Env == Production {
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		db.SetConnMaxIdleTime(10 * time.Minute)
		db.SetConnMaxLifetime(1 * time.Hour)
	} else {
		db.SetMaxOpenConns(2)
		db.SetMaxIdleConns(2)
		db.SetConnMaxIdleTime(2 * time.Minute)
		db.SetConnMaxLifetime(1 * time.Hour)
	}

	//fmt.Println("connecting to ", mysqlConfig.FormatDSN())
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func (ic *Impart) GetMigrationDBConnection() (*sql.DB, error) {

	mysqlConfig := getDefaultConfig()
	mysqlConfig.Addr = fmt.Sprintf("%s:%v", ic.DBHost, ic.DBPort)
	mysqlConfig.User = ic.DBMigrationUsername
	mysqlConfig.Passwd = ic.DBMigrationPassword
	mysqlConfig.DBName = ic.DBName

	db, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		return nil, err
	}
	if ic.Env == Production {
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		db.SetConnMaxIdleTime(10 * time.Minute)
		db.SetConnMaxLifetime(1 * time.Hour)
	} else {
		db.SetMaxOpenConns(2)
		db.SetMaxIdleConns(2)
		db.SetConnMaxIdleTime(2 * time.Minute)
		db.SetConnMaxLifetime(1 * time.Hour)
	}

	//fmt.Println("connecting to ", mysqlConfig.FormatDSN())
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
