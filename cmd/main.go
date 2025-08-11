package main

import (
	"context"
	"crypto/tls"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/jmoiron/sqlx"
	httpSwagger "github.com/swaggo/http-swagger"
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
// @version 1.0.0
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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
		// загружаем сертификаты
		cer, err := tls.LoadX509KeyPair(cfg.SslSert, cfg.SslKey)
		if err != nil {
			logger.Panic("failed certificate loading: %s", zap.Error(err))
		}
		// создаём конфигурацию TLS сервера
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
		// создаём слушателя https соединения
		ln, err := tls.Listen("tcp", ":8080", tlsConfig)
		if err != nil {
			logger.Panic("failed TLS listener creating: %s", zap.Error(err))
		}
		// запускаем веб-сервер с новым TLS слушателем
		err = server.App.Listener(ln)
		if err != nil {
			logger.Panic("http server error: %s", zap.Error(err))
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
	server.App.Use("/swagger/*", web.HTTPHandler(httpSwagger.WrapHandler))
	server.App.Use(requestid.New())
	server.App.Use(recover.New())
	server.GroupApiV1.Use(web.AuthMiddleware(logger))
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
