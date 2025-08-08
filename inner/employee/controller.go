package employee

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"idm/inner/common"
	"idm/inner/web"
	"slices"
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
	GetPage(request PageRequest) (PageResponse, error)
	GetKeySetPage(request PageKeySetRequest) (PageKeySetResponse, error)
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
	c.server.GroupApiV1.Get("/employees/page", c.GetPage)
	c.server.GroupApiV1.Get("/employees/page-key-set", c.GetPage)
	c.server.GroupApiV1.Get("/employees/:id", c.FindById)
	c.server.GroupApiV1.Get("/employees", c.GetAll)
	c.server.GroupApiV1.Post("/employees/search", c.GetGroupById)
	c.server.GroupApiV1.Delete("/employees/batch-delete", c.DeleteGroup)
	c.server.GroupApiV1.Delete("/employees/:id", c.Delete)
}

// CreateEmployee godoc
// @Summary      Create new employee
// @Description  Creates a new employee based on the provided name.
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        request  body      NameRequest  true  "Employee name payload"
// @Success      200      {object}  Response  "ID of created employee"
// @Failure      400      {object}  Response  "Bad request - validation or already exists error"
// @Failure      500      {object}  Response  "Internal server error"
// @Router       /employees [post]
// @Security BearerAuth
func (c *Controller) CreateEmployee(ctx fiber.Ctx) error {
	token := ctx.Locals(web.JwtKey).(*jwt.Token)
	claims := token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmAdmin) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
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

// FindById godoc
// @Summary      Get employee by ID
// @Description  Получает одного сотрудника по ID
// @Tags         employees
// @Param        id path int true "Employee ID"
// @Success      200 {object} Response
// @Failure      400 {object} Response
// @Failure      404 {object} Response
// @Failure      500 {object} Response
// @Router       /employees/{id} [get]
// @Security BearerAuth
func (c *Controller) FindById(ctx fiber.Ctx) error {
	token := ctx.Locals(web.JwtKey).(*jwt.Token)
	claims := token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmAdmin) || !slices.Contains(claims.RealmAccess.Roles, web.IdmUser) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
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

// GetAll godoc
// @Summary      Get all employees
// @Description  Возвращает список всех сотрудников
// @Tags         employees
// @Success      200 {array} Response
// @Failure      500 {object} Response
// @Router       /employees [get]
// @Security BearerAuth
func (c *Controller) GetAll(ctx fiber.Ctx) error {
	token := ctx.Locals(web.JwtKey).(*jwt.Token)
	claims := token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmAdmin) || !slices.Contains(claims.RealmAccess.Roles, web.IdmUser) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
	myCxt := ctx.Context()
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

// GetGroupById godoc
// @Summary      Get employees by IDs
// @Description  Получает сотрудников по списку ID
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        request body IdsRequest true "IDs"
// @Success      200 {array} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /employees/search [post]
// @Security BearerAuth
func (c *Controller) GetGroupById(ctx fiber.Ctx) error {
	token := ctx.Locals(web.JwtKey).(*jwt.Token)
	claims := token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmAdmin) || !slices.Contains(claims.RealmAccess.Roles, web.IdmUser) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
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

// Delete godoc
// @Summary      Delete employee by ID
// @Description  Удаляет сотрудника по ID
// @Tags         employees
// @Param        id path int true "Employee ID"
// @Success      200 "Deleted"
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /employees/{id} [delete]
// @Security BearerAuth
func (c *Controller) Delete(ctx fiber.Ctx) error {
	token := ctx.Locals(web.JwtKey).(*jwt.Token)
	claims := token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmAdmin) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
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

// DeleteGroup godoc
// @Summary      Delete multiple employees by IDs
// @Description  Удаляет нескольких сотрудников по списку ID
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        request body IdsRequest true "IDs"
// @Success      200 "Batch deleted"
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /employees/batch-delete [delete]
// @Security BearerAuth
func (c *Controller) DeleteGroup(ctx fiber.Ctx) error {
	token := ctx.Locals(web.JwtKey).(*jwt.Token)
	claims := token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmAdmin) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
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

// GetPage godoc
// @Summary      Get paginated employees (offset-based)
// @Description  Возвращает сотрудников с пагинацией по номеру страницы
// @Tags         employees
// @Param        pageNumber query int true "Page number"
// @Param        pageSize query int true "Page size"
// @Param        textFilter query string false "Filter by name"
// @Success      200 {array} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /employees/page [get]
// @Security BearerAuth
func (c *Controller) GetPage(ctx fiber.Ctx) error {
	token := ctx.Locals(web.JwtKey).(*jwt.Token)
	claims := token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmAdmin) || !slices.Contains(claims.RealmAccess.Roles, web.IdmUser) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
	number, err := strconv.ParseInt(ctx.Query("pageNumber", "0"), 10, 64)
	if err != nil {
		c.logger.Error("get page of employee: wrong pageNumber", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	size, err := strconv.ParseInt(ctx.Query("pageSize"), 10, 64)
	if err != nil {
		c.logger.Error("get page of employee: wrong pageSize", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	name := ctx.Query("textFilter")
	request := PageRequest{
		PageSize:   size,
		PageNumber: number,
		TextFilter: name,
	}
	employees, err := c.service.GetPage(request)
	var reqErr *common.RequestValidationError
	var notFoundErr *common.NotFoundError
	if err != nil {
		c.logger.Error("get page of employees", zap.Error(err))
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

// GetKeySetPage godoc
// @Summary      Get keyset paginated employees
// @Description  Возвращает сотрудников с пагинацией по ID (keyset)
// @Tags         employees
// @Param        lastId query int true "Last ID"
// @Param        pageSize query int true "Page size"
// @Success      200 {array} Response
// @Failure      400 {object} Response
// @Failure      500 {object} Response
// @Router       /employees/page-key-set [get]
// @Security BearerAuth
func (c *Controller) GetKeySetPage(ctx fiber.Ctx) error {
	token := ctx.Locals(web.JwtKey).(*jwt.Token)
	claims := token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmAdmin) || !slices.Contains(claims.RealmAccess.Roles, web.IdmUser) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
	lastId, err := strconv.ParseInt(ctx.Query("lastId", "1"), 10, 64)
	if err != nil {
		c.logger.Error("get page of employee: wrong id", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	size, err := strconv.ParseInt(ctx.Query("pageSize"), 10, 64)
	if err != nil {
		c.logger.Error("get page of employee: wrong pageSize", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	request := PageKeySetRequest{
		LastId:   lastId,
		PageSize: size,
		IsNext:   true,
	}
	employees, err := c.service.GetKeySetPage(request)
	var reqErr *common.RequestValidationError
	var notFoundErr *common.NotFoundError
	if err != nil {
		c.logger.Error("get page of employees", zap.Error(err))
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
