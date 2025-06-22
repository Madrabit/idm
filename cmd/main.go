package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"idm/inner/common"
	database2 "idm/inner/database"
	"idm/inner/employee"
	"idm/inner/info"
	"idm/inner/role"
	"idm/inner/validator"
	"idm/inner/web"
)

func main() {
	cfg := common.GetConfig(".env")
	db := database2.ConnectDbWithCfg(cfg)
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Printf("error closing db: %v", err)
		}
	}()
	server := build(cfg, db)
	if err := server.App.Listen(":8080"); err != nil {
		panic(fmt.Sprintf("http server error: %s", err))
	}
}

func build(cfg common.Config, database *sqlx.DB) *web.Server {
	server := web.NewServer()
	vld := validator.New()
	employeeRepo := employee.NewRepository(database)
	employeeService := employee.NewService(employeeRepo, vld)
	employeeController := employee.NewController(server, employeeService)
	employeeController.RegisterRoutes()
	roleRepo := role.NewRepository(database)
	roleService := role.NewService(roleRepo, vld)
	roleController := role.NewController(server, roleService)
	roleController.RegisterRoutes()
	infoController := info.NewController(server, cfg, database)
	infoController.RegisterRoutes()
	return server
}
