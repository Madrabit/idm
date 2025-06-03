package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"idm/inner/common"
	"time"
)

func ConnectDb() *sqlx.DB {
	cfg := common.GetConfig(".env")
	return ConnectDbWithCfg(cfg)
}

func ConnectDbWithCfg(cfg common.Config) *sqlx.DB {
	db := sqlx.MustConnect(cfg.DbDriverName, cfg.Dsn)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)
	return db
}
