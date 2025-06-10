package tests

import (
	"context"
	"github.com/jmoiron/sqlx"
	"idm/inner/common"
	"idm/inner/database"
	"idm/inner/employee"
	"log"
)

const env = ".env"

type Fixture struct {
	db        *sqlx.DB
	employees *employee.Repository
}

func NewFixture() *Fixture {
	cfg := common.GetConfig(env)
	db := database.ConnectDbWithCfg(cfg)
	repo := employee.NewRepository(db)
	initSchema(db)
	return &Fixture{db: db, employees: repo}
}

func (f *Fixture) Employee(name string) (int64, error) {
	tx, err := f.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return -1, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	entity := employee.Entity{Name: name}
	return f.employees.Add(tx, entity)
}

func (f *Fixture) Close() {
	err := f.db.Close()
	if err != nil {
		return
	}
}

func (f *Fixture) ClearTable() {
	f.db.MustExec("DELETE FROM employee;")
}

func initSchema(db *sqlx.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS employee
	(
		id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		name       TEXT        NOT NULL,
		created_at timestamptz NOT NULL DEFAULT now(),
		updated_at timestamptz          DEFAULT now()
	);`
	_, err := db.Exec(schema)
	if err != nil {
		log.Fatal("create temp table employee %w", err)
	}
}
