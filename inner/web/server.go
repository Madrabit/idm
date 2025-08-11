package web

import (
	"github.com/gofiber/fiber/v3"

	_ "idm/docs"
)

type Server struct {
	App           *fiber.App
	GroupApiV1    fiber.Router
	GroupInternal fiber.Router
}

type AuthMiddlewareInterface interface {
	ProtectWithJwt() func(*fiber.Ctx) error
}

func NewServer() *Server {
	app := fiber.New()

	groupApi := app.Group("/api")
	groupApiV1 := groupApi.Group("/v1")
	groupInternal := groupApi.Group("/internal")
	return &Server{
		App:           app,
		GroupApiV1:    groupApiV1,
		GroupInternal: groupInternal,
	}
}
