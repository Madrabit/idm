package employee

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"
	"time"
)

type Controller struct {
	server  *web.Server
	service Svc
	logger  *common.Logger
}

type Svc interface {
	FindById(id IdRequest) (employee Response, err error)
	GetAll(ctx context.Context) ([]Response, error)
	Add(request NameRequest) (id int64, err error)
	GetGroupById(ids IdsRequest) ([]Response, error)
	Delete(id IdRequest) error
	DeleteGroup(ids IdsRequest) error
}

func NewController(server *web.Server, service Svc, logger *common.Logger) *Controller {
	return &Controller{
		server:  server,
		service: service,
		logger:  logger,
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
		c.logger.Error("create employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.Debug("create employee: received request", zap.Any("request", request))
	newEmployeeId, err := c.service.Add(request)
	var reqErr *common.RequestValidationError
	var existsErr *common.AlreadyExistsError
	if err != nil {
		c.logger.Error("create employee", zap.Error(err))
		if errors.As(err, &reqErr) || errors.As(err, &existsErr) {
			c.logger.Error("create employee", zap.Error(err))
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		}
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	c.logger.Info("employee created", zap.Int64("id", newEmployeeId))
	return common.OkResponse(ctx, newEmployeeId)
}

func (c *Controller) FindById(ctx fiber.Ctx) error {
	param := ctx.Params("id")
	request, err := strconv.Atoi(param)
	c.logger.Debug("find by id employee: received request", zap.Any("request", request))
	if err != nil {
		c.logger.Error("find by id employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	id := IdRequest{Id: int64(request)}
	employee, err := c.service.FindById(id)
	var reqErr *common.RequestValidationError
	var notFoundErr *common.NotFoundError
	if err != nil {
		c.logger.Error("find by id employee", zap.Error(err))
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
	myCxt := ctx.Context()
	myCtx := context.WithValue(myCxt, "ctxLogger", c.logger)
	ctxLogger := myCtx.Value("ctxLogger").(*zap.Logger)
	ctxLogger.Info("get all employees from Context logger")
	timeoutCtx, cancel := context.WithTimeout(myCxt, time.Second*5)
	defer cancel()
	employees, err := c.service.GetAll(timeoutCtx)
	var notFoundErr *common.NotFoundError
	if err != nil {
		c.logger.Error("get all employees", zap.Error(err))
		switch {
		case errors.As(err, &notFoundErr):
			return common.ErrResponse(ctx, fiber.StatusOK, err.Error())
		case errors.Is(err, context.DeadlineExceeded):
			return common.ErrResponse(ctx, fiber.StatusGatewayTimeout, "timeout exceeded")
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	return common.OkResponse(ctx, employees)
}

func (c *Controller) GetGroupById(ctx fiber.Ctx) error {
	var request IdsRequest
	if err := ctx.Bind().Body(&request); err != nil {
		c.logger.Error("get employees by ids", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	employees, err := c.service.GetGroupById(request)
	var reqErr *common.RequestValidationError
	var notFoundErr *common.NotFoundError
	if err != nil {
		c.logger.Error("get employees by ids", zap.Error(err))
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
	c.logger.Debug("delete employee: received request", zap.Any("request", request))
	if err != nil {
		c.logger.Error("delete employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	id := IdRequest{Id: int64(request)}
	err = c.service.Delete(id)
	var reqErr *common.RequestValidationError
	if err != nil {
		c.logger.Error("delete employee", zap.Error(err))
		switch {
		case errors.As(err, &reqErr):
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	c.logger.Info("employee deleted", zap.Int64("id", int64(request)))
	return nil
}

func (c *Controller) DeleteGroup(ctx fiber.Ctx) error {
	var request IdsRequest
	body := ctx.Body()
	if err := json.Unmarshal(body, &request); err != nil {
		c.logger.Error("delete group employees by ids", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.Debug("delete group employees: received request", zap.Any("request", request))
	err := c.service.DeleteGroup(request)
	var reqErr *common.RequestValidationError
	if err != nil {
		c.logger.Error("delete group employees by ids", zap.Error(err))
		switch {
		case errors.As(err, &reqErr):
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	c.logger.Info("employee deleted")
	return ctx.SendStatus(200)
}
