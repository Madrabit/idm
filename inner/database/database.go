package database

import (
	"github.com/jmoiron/sqlx"
	"idm/inner/common"
	"time"
)

var DB *sqlx.DB

func ConnectDb() *sqlx.DB {
	cfg := common.GetConfig(".env")
	return ConnectDbWithCfg(cfg)
}

func ConnectDbWithCfg(cfg common.Config) *sqlx.DB {
	DB = sqlx.MustConnect(cfg.DbDriverName, cfg.Dsn)
	DB.SetMaxIdleConns(5)
	DB.SetMaxOpenConns(20)
	DB.SetConnMaxLifetime(1 * time.Minute)
	DB.SetConnMaxIdleTime(10 * time.Minute)
	return DB
}
