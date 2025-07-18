package role

import (
	"errors"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"
)

type Controller struct {
	server  *web.Server
	service Svc
	logger  *common.Logger
}

type Svc interface {
	FindById(id IdRequest) (role Response, err error)
	GetAll() ([]Response, error)
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
	c.server.GroupApiV1.Get("/roles", c.GetAll)
	c.server.GroupApiV1.Post("/roles", c.CreateRole)
	c.server.GroupApiV1.Get("/roles/:id", c.FindById)
	c.server.GroupApiV1.Post("/roles/search", c.GetGroupById)
	c.server.GroupApiV1.Delete("/roles/batch-delete", c.DeleteGroup)
	c.server.GroupApiV1.Delete("/roles/:id", c.Delete)
}

// CreateRole godoc
// @Summary      Create role
// @Description  Создаёт новую роль по имени
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        request body NameRequest true "Role name"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /roles [post]
func (c *Controller) CreateRole(ctx fiber.Ctx) error {
	var request NameRequest
	if err := ctx.Bind().Body(&request); err != nil {
		c.logger.Error("create role", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.Debug("create role: received request", zap.Any("request", request))
	newRoleId, err := c.service.Add(request)
	var reqErr *common.RequestValidationError
	var existsErr *common.AlreadyExistsError
	if err != nil {
		c.logger.Error("create role", zap.Error(err))
		if errors.As(err, &reqErr) || errors.As(err, &existsErr) {
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		}
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	c.logger.Info("role created", zap.Int64("id", newRoleId))
	return common.OkResponse(ctx, newRoleId)
}

// FindById godoc
// @Summary      Get role by ID
// @Description  Получает роль по её ID
// @Tags         roles
// @Produce      json
// @Param        id path int true "Role ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      404 {object} Response
// @Failure      500 {object} Response
// @Router       /roles/{id} [get]
func (c *Controller) FindById(ctx fiber.Ctx) error {
	param := ctx.Params("id")
	request, err := strconv.Atoi(param)
	c.logger.Debug("find by id role: received request", zap.Any("request", request))
	if err != nil {
		c.logger.Error("find by id role", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	id := IdRequest{Id: int64(request)}
	role, err := c.service.FindById(id)
	var reqErr *common.RequestValidationError
	var notFoundErr *common.NotFoundError
	if err != nil {
		c.logger.Error("find by id role", zap.Error(err))
		switch {
		case errors.As(err, &notFoundErr):
			return common.ErrResponse(ctx, fiber.StatusOK, err.Error())
		case errors.As(err, &reqErr):
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	return common.OkResponse(ctx, role)
}

// GetAll godoc
// @Summary      Get all roles
// @Description  Получает список всех ролей
// @Tags         roles
// @Produce      json
// @Success      200 {object} Response
// @Failure      500 {object} Response
// @Router       /roles [get]
func (c *Controller) GetAll(ctx fiber.Ctx) error {
	roles, err := c.service.GetAll()
	var notFoundErr *common.NotFoundError
	if err != nil {
		c.logger.Error("get all roles", zap.Error(err))
		switch {
		case errors.As(err, &notFoundErr):
			return common.ErrResponse(ctx, fiber.StatusOK, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	return common.OkResponse(ctx, roles)
}

// GetGroupById godoc
// @Summary      Get roles by IDs
// @Description  Получает список ролей по переданным ID
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        request body IdsRequest true "IDs"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /roles/search [post]
func (c *Controller) GetGroupById(ctx fiber.Ctx) error {
	var request IdsRequest
	if err := ctx.Bind().Body(&request); err != nil {
		c.logger.Error("get roles by ids", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	roles, err := c.service.GetGroupById(request)
	var reqErr *common.RequestValidationError
	var notFoundErr *common.NotFoundError
	if err != nil {
		c.logger.Error("get roles by ids", zap.Error(err))
		switch {
		case errors.As(err, &notFoundErr):
			return common.ErrResponse(ctx, fiber.StatusOK, err.Error())
		case errors.As(err, &reqErr):
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	return common.OkResponse(ctx, roles)
}

// Delete godoc
// @Summary      Delete role by ID
// @Description  Удаляет роль по ID
// @Tags         roles
// @Produce      json
// @Param        id path int true "Role ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /roles/{id} [delete]
func (c *Controller) Delete(ctx fiber.Ctx) error {
	param := ctx.Params("id")
	request, err := strconv.Atoi(param)
	c.logger.Debug("delete role: received request", zap.Any("request", request))
	if err != nil {
		c.logger.Error("delete role", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	id := IdRequest{Id: int64(request)}
	err = c.service.Delete(id)
	var reqErr *common.RequestValidationError
	if err != nil {
		c.logger.Error("delete role", zap.Error(err))
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

// DeleteGroup godoc
// @Summary      Delete roles by IDs
// @Description  Удаляет роли по списку ID
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        request body IdsRequest true "IDs"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /roles/batch-delete [delete]
func (c *Controller) DeleteGroup(ctx fiber.Ctx) error {
	var request IdsRequest
	if err := ctx.Bind().Body(&request); err != nil {
		c.logger.Error("delete group roles by ids", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	err := c.service.DeleteGroup(request)
	var reqErr *common.RequestValidationError
	if err != nil {
		c.logger.Error("delete group roles by ids", zap.Error(err))
		switch {
		case errors.As(err, &reqErr):
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	c.logger.Info("roles deleted")
	return nil
}
