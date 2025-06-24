package info

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"idm/inner/common"
	"idm/inner/web"
	"log"
	"net/http/httptest"
	"testing"
)

func setupTestApp(db *sql.DB, cfg common.Config) *fiber.App {
	app := fiber.New()
	server := &web.Server{
		App:           app,
		GroupInternal: app.Group("/internal"),
	}
	newDb := sqlx.NewDb(db, "sqlmock")
	ctrl := NewController(server, cfg, newDb, nil)
	ctrl.db = newDb
	ctrl.RegisterRoutes()
	return app
}

func TestGetInfo(t *testing.T) {
	cfg := common.Config{
		AppName:    "Idm",
		AppVersion: "0.0.0",
	}
	app := setupTestApp(nil, cfg)
	req := httptest.NewRequest("GET", "/internal/info", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetHealth_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		err := db.Close()
		if err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()
	mock.ExpectClose()
	cfg := common.Config{}
	app := setupTestApp(db, cfg)
	req := httptest.NewRequest("GET", "/internal/health", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetHealth_Down(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer func() {
		err := db.Close()
		if err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()
	mock.ExpectPing().WillReturnError(errors.New("db is down"))
	mock.ExpectClose()
	cfg := common.Config{}
	app := setupTestApp(db, cfg)
	req := httptest.NewRequest("GET", "/internal/health", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 503, resp.StatusCode)
}
