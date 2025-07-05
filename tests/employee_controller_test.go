package tests

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"idm/inner/common"
	"idm/inner/database"
	"idm/inner/employee"
	"idm/inner/validator"
	"idm/inner/web"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestPaginationIntegration(t *testing.T) {
	cfg := common.GetConfig(".env")
	logger := common.NewLogger(cfg)
	defer func() { _ = logger.Sync() }()
	db := database.ConnectDbWithCfg(cfg)
	db.MustExec("DELETE FROM employee;")
	app := web.NewServer()
	vld := validator.New()
	employeeRepo := employee.NewRepository(db)
	employeeService := employee.NewService(employeeRepo, vld)
	employeeController := employee.NewController(app, employeeService, logger)
	employeeController.RegisterRoutes()
	for i := 1; i <= 5; i++ {
		_, err := employeeService.Add(employee.NameRequest{Name: "name" + strconv.Itoa(i)})
		if err != nil {
			logger.Error("error adding in test employee controller: %v", zap.Error(err))
		}
	}
	a := assert.New(t)
	t.Run("Page 1: 3 items", func(t *testing.T) {
		url := "/api/v1/employees/page?pageNumber=0&pageSize=3"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		resp, err := app.App.Test(req)
		a.NoError(err)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println("error closing body")
			}
		}(resp.Body)
		a.Equal(http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		a.NoError(err)
		result := common.Response[employee.PageResponse]{}
		err = json.Unmarshal(body, &result)
		a.NoError(err)
		a.True(result.Success)
		a.Len(result.Data.Result, 3)
		a.Equal(int64(3), result.Data.PageSize)
		a.Equal(int64(0), result.Data.PageNumber)
		a.Equal(int64(5), result.Data.Total)
		defer func() {
			if err := db.Close(); err != nil {
				fmt.Println("error closing db")
			}
		}()
	})

	t.Run("Page 2: 2 items", func(t *testing.T) {
		url := "/api/v1/employees/page?pageNumber=1&pageSize=3"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		resp, err := app.App.Test(req)
		a.NoError(err)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println("error closing body")
			}
		}(resp.Body)
		a.Equal(http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		a.NoError(err)
		result := common.Response[employee.PageResponse]{}
		err = json.Unmarshal(body, &result)
		a.NoError(err)
		a.True(result.Success)
		a.Len(result.Data.Result, 2)
		a.Equal(int64(3), result.Data.PageSize)
		a.Equal(int64(1), result.Data.PageNumber)
		a.Equal(int64(5), result.Data.Total)
		defer func() {
			if err := db.Close(); err != nil {
				fmt.Println("error closing db")
			}
		}()
	})

	t.Run("Page 3: 0 items", func(t *testing.T) {
		url := "/api/v1/employees/page?pageNumber=3&pageSize=3"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		resp, err := app.App.Test(req)
		a.NoError(err)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println("error closing body")
			}
		}(resp.Body)
		a.Equal(http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		a.NoError(err)
		result := common.Response[employee.PageResponse]{}
		err = json.Unmarshal(body, &result)
		a.NoError(err)
		a.True(result.Success)
		a.Len(result.Data.Result, 0)
		a.Equal(int64(3), result.Data.PageSize)
		a.Equal(int64(3), result.Data.PageNumber)
		a.Equal(int64(5), result.Data.Total)
		defer func() {
			if err := db.Close(); err != nil {
				fmt.Println("error closing db")
			}
		}()

	})

	t.Run("Invalid query: negative pageNumber", func(t *testing.T) {
		url := "/api/v1/employees/page?pageNumber=-1&pageSize=3"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		resp, err := app.App.Test(req)
		a.NoError(err)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println("error closing body")
			}
		}(resp.Body)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		a.NoError(err)
		result := common.Response[employee.PageResponse]{}
		err = json.Unmarshal(body, &result)
		a.NoError(err)
		a.False(result.Success)
		a.Contains(result.Message, "validation") // ожидаем текст ошибки
	})
	t.Run("Missing pageNumber", func(t *testing.T) {
		// pageNumber по умолчанию = 0 в контроллере
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/page?pageSize=3", nil)
		resp, err := app.App.Test(req)
		a.NoError(err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("Missing pageSize", func(t *testing.T) {
		// вернуть ошибку 400
		// непредсказуемый размер страницы может привести к нагрузке на БД
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/page?pageNumber=0", nil)
		resp, err := app.App.Test(req)
		a.NoError(err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
