package employee

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v3"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"
)

type Controller struct {
	server  *web.Server
	service Svc
}

type Svc interface {
	FindById(id IdRequest) (employee Response, err error)
	GetAll() ([]Response, error)
	Add(request NameRequest) (id int64, err error)
	GetGroupById(ids IdsRequest) ([]Response, error)
	Delete(id IdRequest) error
	DeleteGroup(ids IdsRequest) error
}

func NewController(server *web.Server, service Svc) *Controller {
	return &Controller{
		server:  server,
		service: service,
	}
}

func (c *Controller) RegisterRoutes() {
	c.server.GroupApiV1.Post("/employees", c.CreateEmployee)
	c.server.GroupApiV1.Get("/employees/:id", c.FindById)
	c.server.GroupApiV1.Get("/employees", c.GetAll)
	c.server.GroupApiV1.Post("/employees/search", c.GetGroupById)
	c.server.GroupApiV1.Delete("/employees/batch-delete", c.DeleteGroup)
	c.server.GroupApiV1.Delete("/employees/:id", c.Delete)
}

func (c *Controller) CreateEmployee(ctx fiber.Ctx) error {
	var request NameRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	newEmployeeId, err := c.service.Add(request)
	var reqErr *common.RequestValidationError
	var existsErr *common.AlreadyExistsError
	if err != nil {
		if errors.As(err, &reqErr) || errors.As(err, &existsErr) {
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		}
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, newEmployeeId)
}

func (c *Controller) FindById(ctx fiber.Ctx) error {
	param := ctx.Params("id")
	request, err := strconv.Atoi(param)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	id := IdRequest{Id: int64(request)}
	employee, err := c.service.FindById(id)
	var reqErr *common.RequestValidationError
	var notFoundErr *common.NotFoundError
	if err != nil {
		switch {
		case errors.As(err, &notFoundErr):
			return common.ErrResponse(ctx, fiber.StatusOK, err.Error())
		case errors.As(err, &reqErr):
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	return common.OkResponse(ctx, employee)
}

func (c *Controller) GetAll(ctx fiber.Ctx) error {
	employees, err := c.service.GetAll()
	var notFoundErr *common.NotFoundError
	if err != nil {
		switch {
		case errors.As(err, &notFoundErr):
			return common.ErrResponse(ctx, fiber.StatusOK, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	return common.OkResponse(ctx, employees)
}

func (c *Controller) GetGroupById(ctx fiber.Ctx) error {
	var request IdsRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	employees, err := c.service.GetGroupById(request)
	var reqErr *common.RequestValidationError
	var notFoundErr *common.NotFoundError
	if err != nil {
		switch {
		case errors.As(err, &notFoundErr):
			return common.ErrResponse(ctx, fiber.StatusOK, err.Error())
		case errors.As(err, &reqErr):
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	return common.OkResponse(ctx, employees)
}

func (c *Controller) Delete(ctx fiber.Ctx) error {
	param := ctx.Params("id")
	request, err := strconv.Atoi(param)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	id := IdRequest{Id: int64(request)}
	err = c.service.Delete(id)
	var reqErr *common.RequestValidationError
	if err != nil {
		switch {
		case errors.As(err, &reqErr):
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	return nil
}

func (c *Controller) DeleteGroup(ctx fiber.Ctx) error {
	var request IdsRequest
	body := ctx.Body()
	if err := json.Unmarshal(body, &request); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	err := c.service.DeleteGroup(request)
	var reqErr *common.RequestValidationError
	if err != nil {
		switch {
		case errors.As(err, &reqErr):
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	return ctx.SendStatus(200)
}
