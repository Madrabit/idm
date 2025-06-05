package tests

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"idm/inner/common"
	"idm/inner/database"
	Role "idm/inner/role"
	"log"
)

type RoleFixture struct {
	db   *sqlx.DB
	repo *Role.Repository
}

func NewRoleFixture() *RoleFixture {
	cfg := common.GetConfig(env)
	db := database.ConnectDbWithCfg(cfg)
	repo := Role.NewRoleRepository(db)
	initRoleSchema(db)
	return &RoleFixture{db, repo}
}

func (f *RoleFixture) Role(name string) (int64, error) {
	entity := Role.Entity{Name: name}
	newId, err := f.repo.Add(entity)
	if err != nil {
		return -1, fmt.Errorf("fall while add role: %w", err)
	}
	return newId, nil
}

func (f *RoleFixture) Close() {
	err := f.db.Close()
	if err != nil {
		return
	}
}

func (f *RoleFixture) ClearTable() {
	f.db.MustExec("DELETE FROM role;")
}

func initRoleSchema(db *sqlx.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS role
	(
		id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		name       TEXT        NOT NULL,
		created_at timestamptz NOT NULL DEFAULT now(),
		updated_at timestamptz          DEFAULT now()
	);`
	_, err := db.Exec(schema)
	if err != nil {
		log.Fatal("create temp table role: %w", err)
	}
}
