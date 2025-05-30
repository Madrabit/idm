package tests

import (
	_ "github.com/lib/pq"
	"idm/inner/common"
	"idm/inner/database"
	"testing"
)

func TestDatabaseConnection(t *testing.T) {
	t.Run("wrong DB_DSN", func(t *testing.T) {
		t.Setenv("DB_DSN", "'host=127.0.0.1 port=1111 user=1 password=1 dbname=test_idm sslmode=disable'")
		config := common.GetConfig("")
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("MustConnect do not panic")
			}
		}()
		database.ConnectDbWithCfg(config)
	})
	t.Run("connection to db works", func(t *testing.T) {
		t.Setenv("DB_DRIVER_NAME", "postgres")
		t.Setenv("DB_DSN", "host=127.0.0.1 port=5434 user=admin password=postgres dbname=test_idm sslmode=disable")
		config := common.GetConfig("")
		defer func() {
			if r := recover(); r != nil {
				t.Fatal("MustConnect panic")
			}
		}()
		database.ConnectDbWithCfg(config)
	})

}
