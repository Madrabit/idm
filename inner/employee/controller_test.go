package employee

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"idm/inner/common"
	"idm/inner/web"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

type MockService struct {
	mock.Mock
}

func (svc *MockService) FindById(id IdRequest) (Response, error) {
	args := svc.Called(id)
	return args.Get(0).(Response), args.Error(1)
}

func (svc *MockService) Add(request NameRequest) (int64, error) {
	args := svc.Called(request)
	return args.Get(0).(int64), args.Error(1)
}

func (svc *MockService) GetAll() ([]Response, error) {
	args := svc.Called()
	return args.Get(0).([]Response), args.Error(1)
}

func (svc *MockService) GetGroupById(ids IdsRequest) ([]Response, error) {
	args := svc.Called(ids)
	return args.Get(0).([]Response), args.Error(1)
}

func (svc *MockService) Delete(id IdRequest) error {
	args := svc.Called(id)
	return args.Error(0)
}

func (svc *MockService) DeleteGroup(ids IdsRequest) error {
	args := svc.Called(ids)
	return args.Error(0)
}

func TestController_Add(t *testing.T) {
	var a = assert.New(t)
	t.Run("should return created employee id", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, nil)
		controller.RegisterRoutes()
		body := strings.NewReader("{\"name\": \"john doe\"}")
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees", body)
		req.Header.Set("Content-Type", "application/json")
		svc.On("Add", mock.AnythingOfType("NameRequest")).Return(int64(123), nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.NotEmpty(resp)
		a.Equal(http.StatusOK, resp.StatusCode)
		bytesData, err := io.ReadAll(resp.Body)
		a.Nil(err)
		var responseBody common.Response[int64]
		err = json.Unmarshal(bytesData, &responseBody)
		a.Nil(err)
		a.Equal(int64(123), responseBody.Data)
		a.True(responseBody.Success)
		a.Empty(responseBody.Message)
	})
	t.Run("should return 400 if request is invalid (RequestValidationError)", func(t *testing.T) {
		var a = assert.New(t)
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, nil)
		controller.RegisterRoutes()
		body := strings.NewReader("{\"name\": \"\"}")
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees", body)
		req.Header.Set("Content-Type", "application/json")
		svc.On("Add", mock.AnythingOfType("NameRequest")).Return(int64(0), &common.RequestValidationError{Massage: "IDs are required"})
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})
	t.Run("should return 409 if employee already exists (AlreadyExistsError)", func(t *testing.T) {
		var a = assert.New(t)
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, nil)
		controller.RegisterRoutes()
		body := strings.NewReader(`{"name": "John"}`)
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees", body)
		req.Header.Set("Content-Type", "application/json")
		svc.On("Add", mock.AnythingOfType("NameRequest")).Return(int64(0), &common.AlreadyExistsError{Massage: "employee already exists"})
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode) // 409
	})
}
func TestController_FindById(t *testing.T) {
	var a = assert.New(t)
	t.Run("should retrieve employee by id", func(t *testing.T) {
		server := web.NewServer()
		var svc = new(MockService)
		var controller = NewController(server, svc, nil)
		controller.RegisterRoutes()
		fixedTime := time.Date(2025, time.June, 17, 20, 19, 30, 0, time.UTC)
		employee := Response{
			Id:       1,
			Name:     "john doe",
			CreateAt: fixedTime,
			UpdateAt: fixedTime,
		}
		var req = httptest.NewRequest("GET", "/api/v1/employees/1", nil)
		req.Header.Set("Content-Type", "application/json")
		svc.On("FindById", IdRequest{1}).Return(employee, nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.NotEmpty(resp)
		a.Equal(http.StatusOK, resp.StatusCode)
	})
	t.Run("should return 400 if request is invalid (RequestValidationError)", func(t *testing.T) {
		var a = assert.New(t)
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, nil)
		controller.RegisterRoutes()
		svc.On("FindById", IdRequest{int64(0)}).Return(Response{}, &common.RequestValidationError{Massage: "ID are required"})
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/0", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})
	t.Run("should return 200 if request is invalid (NotFoundError)", func(t *testing.T) {
		var a = assert.New(t)
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, nil)
		controller.RegisterRoutes()
		svc.On("FindById", IdRequest{int64(1)}).Return(Response{}, &common.NotFoundError{Massage: "not found employee"})
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/1", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusOK, resp.StatusCode)
	})
}
func TestController_GetAll(t *testing.T) {
	var a = assert.New(t)
	t.Run("should return all employees", func(t *testing.T) {
		server := web.NewServer()
		var svc = new(MockService)
		var controller = NewController(server, svc, nil)
		controller.RegisterRoutes()
		var req = httptest.NewRequest("GET", "/api/v1/employees", nil)
		req.Header.Set("Content-Type", "application/json")
		fixedTime := time.Date(2025, time.June, 17, 20, 19, 30, 0, time.UTC)
		entity := []Response{
			{Id: 1,
				Name:     "john doe",
				CreateAt: fixedTime,
				UpdateAt: fixedTime,
			},
			{Id: 2,
				Name:     "Ivan Ivan",
				CreateAt: fixedTime,
				UpdateAt: fixedTime,
			},
		}
		svc.On("GetAll").Return(entity, nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.NotEmpty(resp)
		a.Equal(http.StatusOK, resp.StatusCode)
		bytesData, err := io.ReadAll(resp.Body)
		a.Nil(err)
		var responseBody common.Response[[]Response]
		err = json.Unmarshal(bytesData, &responseBody)
		a.Nil(err)
		a.Equal(entity, responseBody.Data)
		a.True(responseBody.Success)
		a.Empty(responseBody.Message)
	})
	t.Run("should return 200 if request is invalid (NotFoundError)", func(t *testing.T) {
		var a = assert.New(t)
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, nil)
		controller.RegisterRoutes()
		svc.On("GetAll", mock.Anything).Return([]Response{}, &common.NotFoundError{Massage: "not found employee"})
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusOK, resp.StatusCode)
	})
}
func TestController_GetGroupById(t *testing.T) {
	var a = assert.New(t)
	t.Run("should return employees by ids", func(t *testing.T) {
		server := web.NewServer()
		var svc = new(MockService)
		var controller = NewController(server, svc, nil)
		controller.RegisterRoutes()
		request := IdsRequest{Ids: []int64{1, 2}}
		marshal, err := json.Marshal(request)
		a.Nil(err)
		body := bytes.NewReader(marshal)
		req := httptest.NewRequest("POST", "/api/v1/employees/search", body)
		req.Header.Set("Content-Type", "application/json")
		fixedTime := time.Date(2025, time.June, 17, 20, 19, 30, 0, time.UTC)
		entity := []Response{
			{Id: 1,
				Name:     "john doe",
				CreateAt: fixedTime,
				UpdateAt: fixedTime,
			},
			{Id: 2,
				Name:     "Ivan Ivan",
				CreateAt: fixedTime,
				UpdateAt: fixedTime,
			},
		}
		svc.On("GetGroupById", request).Return(entity, nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.NotEmpty(resp)
		a.Equal(http.StatusOK, resp.StatusCode)
		bytesData, err := io.ReadAll(resp.Body)
		a.Nil(err)
		var responseBody common.Response[[]Response]
		err = json.Unmarshal(bytesData, &responseBody)
		a.Nil(err)
		a.Equal(entity, responseBody.Data)
		a.True(responseBody.Success)
		a.Empty(responseBody.Message)
	})
	t.Run("should return 400 if request is invalid (RequestValidationError)", func(t *testing.T) {
		var a = assert.New(t)
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, nil)
		controller.RegisterRoutes()
		invalidRequest := IdsRequest{Ids: nil}
		requestBody, err := json.Marshal(invalidRequest)
		a.Nil(err)

		svc.On("GetGroupById", mock.Anything).Return(&common.RequestValidationError{Massage: "IDs are required"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/employees/search", bytes.NewReader(requestBody))
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})
	t.Run("should return 200 if request is invalid (NotFoundError)", func(t *testing.T) {
		var a = assert.New(t)
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, nil)
		controller.RegisterRoutes()
		request := IdsRequest{Ids: []int64{1, 2}}
		marshal, err := json.Marshal(request)
		a.Nil(err)
		body := bytes.NewReader(marshal)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/employees/search", body)
		req.Header.Set("Content-Type", "application/json")
		svc.On("GetGroupById", mock.Anything).Return([]Response{}, &common.NotFoundError{Massage: "not found employee"})
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusOK, resp.StatusCode)
	})
}
func TestController_Delete(t *testing.T) {
	var a = assert.New(t)
	t.Run("should delete employee by id", func(t *testing.T) {
		server := web.NewServer()
		var svc = new(MockService)
		var controller = NewController(server, svc, nil)
		controller.RegisterRoutes()
		var req = httptest.NewRequest("DELETE", "/api/v1/employees/1", nil)
		req.Header.Set("Content-Type", "application/json")
		svc.On("Delete", IdRequest{1}).Return(nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.NotEmpty(resp)
		a.Equal(http.StatusOK, resp.StatusCode)
	})
	t.Run("should return 400 if request is invalid (RequestValidationError)", func(t *testing.T) {
		var a = assert.New(t)
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, nil)
		controller.RegisterRoutes()
		svc.On("Delete", IdRequest{int64(0)}).Return(&common.RequestValidationError{Massage: "ID are required"})
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/employees/0", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})
}
func TestController_DeleteGroup(t *testing.T) {
	var a = assert.New(t)
	t.Run("should delete employees by ids", func(t *testing.T) {
		server := web.NewServer()
		var svc = new(MockService)
		var controller = NewController(server, svc, nil)
		controller.RegisterRoutes()
		request := IdsRequest{Ids: []int64{1, 2}}
		requestBody, err := json.Marshal(request)
		a.Nil(err)
		svc.On("DeleteGroup", mock.MatchedBy(func(req IdsRequest) bool {
			return reflect.DeepEqual(req.Ids, []int64{1, 2})
		})).Return(nil)
		req := httptest.NewRequest(
			http.MethodDelete,
			"/api/v1/employees/batch-delete",
			bytes.NewReader(requestBody),
		)
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req)
		svc.AssertCalled(t, "DeleteGroup", IdsRequest{Ids: []int64{1, 2}})
		a.Nil(err)
		a.NotEmpty(resp)
		a.Equal(http.StatusOK, resp.StatusCode)
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Response body: %s", string(body))
		}
	})
	t.Run("should return 400 if request is invalid (RequestValidationError)", func(t *testing.T) {
		var a = assert.New(t)
		server := web.NewServer()
		svc := new(MockService)
		controller := NewController(server, svc, nil)
		controller.RegisterRoutes()
		invalidRequest := IdsRequest{Ids: nil}
		requestBody, err := json.Marshal(invalidRequest)
		a.Nil(err)
		svc.On("DeleteGroup", mock.Anything).Return(&common.RequestValidationError{Massage: "IDs are required"})
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/employees/batch-delete", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})
}
