package info

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"idm/inner/common"
	"idm/inner/web"
)

type Controller struct {
	server *web.Server
	cfg    common.Config
	db     *sqlx.DB
	logger *common.Logger
}

func NewController(server *web.Server, cfg common.Config, db *sqlx.DB, logger *common.Logger) *Controller {
	return &Controller{
		server: server,
		cfg:    cfg,
		db:     db,
		logger: logger,
	}
}

type Response struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (c *Controller) RegisterRoutes() {
	c.server.GroupInternal.Get("/info", c.GetInfo)
	c.server.GroupInternal.Get("/health", c.GetHealth)
}

func (c *Controller) GetInfo(ctx fiber.Ctx) error {
	var err = ctx.Status(fiber.StatusOK).JSON(&Response{
		Name:    c.cfg.AppName,
		Version: c.cfg.AppVersion,
	})
	if err != nil {
		c.logger.Error("get info", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "error returning info")
	}
	return nil
}

func (c *Controller) GetHealth(ctx fiber.Ctx) error {
	if err := c.db.Ping(); err != nil {
		c.logger.Error("get health", zap.Error(err))
		return ctx.Status(fiber.StatusServiceUnavailable).SendString("DOWN")
	}
	return ctx.Status(fiber.StatusOK).SendString("OK")
}
