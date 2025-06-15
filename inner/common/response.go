package common

import "github.com/gofiber/fiber/v3"

type Response[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"error"`
	Data    T      `json:"data"`
}

func ErrResponse(c fiber.Ctx, code int, msg string) error {
	c.Status(code)
	return c.JSON(&Response[any]{
		Success: false,
		Message: msg,
		Data:    nil,
	})
}

func OkResponse[T any](
	c fiber.Ctx,
	data T,
) error {
	return c.JSON(&Response[T]{
		Success: true,
		Data:    data,
	})
}
