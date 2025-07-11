package main

import (
	"context"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"idm/inner/common"
	"idm/inner/database"
	"idm/inner/employee"
	"idm/inner/info"
	"idm/inner/role"
	"idm/inner/validator"
	"idm/inner/web"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// @title IDM API documentation
// @BasePath /api/v1
func main() {
	cfg := common.GetConfig(".env")
	logger := common.NewLogger(cfg)
	defer func() { _ = logger.Sync() }()
	db := database.ConnectDbWithCfg(cfg)
	defer func() {
		if err := db.Close(); err != nil {
			logger.Panic("error closing db: %v", zap.Error(err))
		}
	}()
	server := build(cfg, db, logger)
	go func() {
		if err := server.App.Listen(":8080"); err != nil {
			logger.Panic("http server error: %v", zap.Error(err))
		}
	}()
	var wg = &sync.WaitGroup{}
	wg.Add(1)
	go gracefulShutdown(server, wg, logger)
	wg.Wait()
	logger.Info("Graceful shutdown complete.")
}

func gracefulShutdown(server *web.Server, wg *sync.WaitGroup, logger *common.Logger) {
	const shutdownTimeout = 5 * time.Second
	defer wg.Done()
	shutdownSignal, unsubscribeSignal := signal.NotifyContext(context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	)
	defer unsubscribeSignal()
	<-shutdownSignal.Done()
	shutdownCtx, clearCtx := context.WithTimeout(context.Background(), shutdownTimeout)
	defer clearCtx()
	if err := server.App.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown with error: %v\n", zap.Error(err))
		return
	}
	logger.Info("Server exiting")
}

func build(cfg common.Config, database *sqlx.DB, logger *common.Logger) *web.Server {
	server := web.NewServer()
	vld := validator.New()
	employeeRepo := employee.NewRepository(database)
	employeeService := employee.NewService(employeeRepo, vld)
	employeeController := employee.NewController(server, employeeService, logger)
	employeeController.RegisterRoutes()
	roleRepo := role.NewRepository(database)
	roleService := role.NewService(roleRepo, vld)
	roleController := role.NewController(server, roleService, logger)
	roleController.RegisterRoutes()
	infoController := info.NewController(server, cfg, database, logger)
	infoController.RegisterRoutes()
	return server
}
