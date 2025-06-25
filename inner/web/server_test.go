package web

import (
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware(t *testing.T) {
	t.Run("server panic", func(t *testing.T) {
		a := assert.New(t)

		server := NewServer()
		server.App.Get("/panic", func(ctx fiber.Ctx) error {
			panic("crash server with panic")
		})
		req, err := http.NewRequest("GET", "/panic", nil)
		a.NoError(err)
		resp, err := server.App.Test(req)
		a.NoError(err)
		a.Equal(fiber.StatusInternalServerError, resp.StatusCode)
	})
	t.Run("server alive after panic", func(t *testing.T) {
		a := assert.New(t)
		server := NewServer()
		server.App.Get("/panic", func(ctx fiber.Ctx) error {
			panic("crash server with panic")
		})
		server.App.Get("/alive", func(ctx fiber.Ctx) error {
			return ctx.SendStatus(http.StatusOK)
		})
		req, err := http.NewRequest("GET", "/panic", nil)
		a.NoError(err)
		resp, err := server.App.Test(req)
		a.NoError(err)
		a.Equal(fiber.StatusInternalServerError, resp.StatusCode)
		reqAlive, err := http.NewRequest("GET", "/alive", nil)
		a.NoError(err)
		respAlive, err := server.App.Test(reqAlive)
		a.NoError(err)
		a.Equal(fiber.StatusOK, respAlive.StatusCode)
	})
}

func TestRequestIDMiddleware(t *testing.T) {
	a := assert.New(t)
	server := NewServer()
	server.App.Get("/test", func(c fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = string(c.Response().Header.Peek("X-Request-ID"))
		}
		return c.SendString(string(requestID))
	})
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := server.App.Test(req)
	a.NoError(err)
	a.Equal(fiber.StatusOK, resp.StatusCode)
	a.NotEmpty(resp.Header.Get("X-Request-ID"))
}
