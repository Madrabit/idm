package common

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetConfig(t *testing.T) {
	t.Run("empty .env - full ram env", func(t *testing.T) {
		_ = eachEnvFile(t, "")
		t.Setenv("DB_DRIVER_NAME", "postgres")
		t.Setenv("DB_DSN", "great dsn string")
		t.Setenv("APP_NAME", "idm")
		t.Setenv("APP_VERSION", "0.0.0")
		t.Setenv("SSL_SERT", "certs/ssl.cert")
		t.Setenv("SSL_KEY", "certs/ssl.key")
		cnf := GetConfig("")
		assert.Equal(t, "postgres", cnf.DbDriverName)
		assert.Equal(t, "great dsn string", cnf.Dsn)
	})
	t.Run("empty .env but full environment", func(t *testing.T) {
		t.Setenv("DB_DRIVER_NAME", "postgres")
		t.Setenv("DB_DSN", "great dsn string")
		t.Setenv("APP_NAME", "idm")
		t.Setenv("APP_VERSION", "0.0.0")
		t.Setenv("SSL_SERT", "certs/ssl.cert")
		t.Setenv("SSL_KEY", "certs/ssl.key")
		file := eachEnvFile(t, "")
		cnf := GetConfig(file)
		assert.NotEqual(t, Config{}, cnf, "config not exists")
		assert.NotEmpty(t, cnf.DbDriverName, "environments not empty")
		assert.NotEmpty(t, cnf.Dsn, "environments not empty")
	})
	t.Run("check env load priority", func(t *testing.T) {
		t.Setenv("DB_DRIVER_NAME", "ram_env")
		t.Setenv("DB_DSN", "ram_env")
		t.Setenv("APP_NAME", "idm")
		t.Setenv("APP_VERSION", "0.0.0")
		t.Setenv("SSL_SERT", "certs/ssl.cert")
		t.Setenv("SSL_KEY", "certs/ssl.key")
		file := eachEnvFile(t, "GOOSE_DBSTRING=mock_DBSTRING\nDB_DRIVER_NAME=postgres\nDB_DSN=mock_DSN\nAPP_NAME=idm\nAPP_VERSION=0.0.0")
		cnf := GetConfig(file)
		assert.Equal(t, "ram_env", cnf.Dsn, "expected values from environment")
		assert.Equal(t, "ram_env", cnf.DbDriverName, "expected values from environment")
	})
}

func eachEnvFile(t *testing.T, str string) string {
	t.Helper()
	temp, err := os.CreateTemp(".", ".env")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Remove(temp.Name())
	})
	_, err = temp.WriteString(str)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := temp.Close(); err != nil {
			t.Errorf("failed to close temp file: %v", err)
		}
	}()
	return temp.Name()
}
