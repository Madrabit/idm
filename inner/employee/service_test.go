package employee

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"idm/inner/common"
	"idm/inner/validator"
	"testing"
	"time"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) FindWithPagination(tx *sqlx.Tx, offset, limit int64) ([]Entity, error) {
	panic("implement me")
}

func (m *MockRepo) GetTotal(tx *sqlx.Tx) (count int64, err error) {
	panic("implement me")
}

func (m *MockRepo) FindById(id int64) (Entity, error) {
	args := m.Called(id)
	return args.Get(0).(Entity), args.Error(1)
}

func (m *MockRepo) GetAll(context.Context) ([]Entity, error) {
	args := m.Called()
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *MockRepo) Add(tx *sqlx.Tx, employee Entity) (int64, error) {
	args := m.Called(tx, employee)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepo) GetGroupById(ids []int64) ([]Entity, error) {
	args := m.Called(ids)
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *MockRepo) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockRepo) DeleteGroup(ids []int64) error {
	args := m.Called(ids)
	return args.Error(0)
}

func (m *MockRepo) BeginTransaction() (tx *sqlx.Tx, err error) {
	args := m.Called()
	return args.Get(0).(*sqlx.Tx), args.Error(1)
}

func (m *MockRepo) FindByNameTx(tx *sqlx.Tx, name string) (bool, error) {
	args := m.Called(tx, name)
	return args.Get(0).(bool), args.Error(1)
}

func TestFindById(t *testing.T) {
	a := assert.New(t)
	t.Run("should return found employee", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		entity := Entity{
			Id:        1,
			Name:      "John",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		want := entity.toResponse()
		repo.On("FindById", int64(1)).Return(entity, nil)
		got, err := srv.FindById(IdRequest{1})
		a.NoError(err)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "FindById", 1))
	})
	t.Run("should return wrapped error", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		entity := Entity{}
		id := int64(1)
		want := &common.NotFoundError{Massage: fmt.Sprintf("employee service: find by id: employee not found: id=%d", id)}
		repo.On("FindById", id).Return(entity, sql.ErrNoRows)
		response, got := srv.FindById(IdRequest{id})
		var notFoundErr *common.NotFoundError
		a.True(errors.As(got, &notFoundErr))
		a.Empty(response)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "FindById", 1))
	})
}

func TestAdd(t *testing.T) {
	a := assert.New(t)
	t.Run("should return wrapped error while begin transaction", func(t *testing.T) {
		a := assert.New(t)
		db, m, err := sqlmock.New()
		a.NoError(err)
		m.ExpectBegin().WillReturnError(fmt.Errorf("error beginning transaction"))
		sqlxDB := sqlx.NewDb(db, "postgres")
		repo := NewRepository(sqlxDB)
		_, err = repo.BeginTransaction()
		a.Error(err)
		a.Equal("error beginning transaction", err.Error())
	})
	t.Run("should return error finding employee by name", func(t *testing.T) {
		a := assert.New(t)
		db, mockDB, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := db.Close()
			if err != nil {
				t.Fatal("error close database")
			}
		}()
		sqlxDB := sqlx.NewDb(db, "sqlmock")
		mockDB.ExpectBegin()
		tx, err := sqlxDB.Beginx()
		mockDB.ExpectClose()
		if err != nil {
			t.Fatal(err)
		}
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		if err != nil {
			t.Fatal(err)
		}
		entity := Entity{
			Id:        1,
			Name:      "John",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.On("BeginTransaction").Return(tx, nil)
		want := fmt.Errorf("rollback failed: original error: employee service: add employee: error checking exists employee")
		repo.On("FindByNameTx", tx, entity.Name).Return(false, want)
		response, got := srv.Add(NameRequest{entity.Name})
		a.Empty(response)
		a.NotNil(got)
		a.ErrorContains(got, want.Error())
		a.True(repo.AssertNumberOfCalls(t, "FindByNameTx", 1))
		a.True(repo.AssertNumberOfCalls(t, "BeginTransaction", 1))

	})
	t.Run("should return employee exists", func(t *testing.T) {
		a := assert.New(t)
		db, mockDB, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := db.Close()
			if err != nil {
				t.Fatal("error close database")
			}
		}()
		sqlxDB := sqlx.NewDb(db, "sqlmock")
		mockDB.ExpectBegin()
		mockDB.ExpectClose()
		tx, err := sqlxDB.Beginx()
		if err != nil {
			t.Fatal(err)
		}
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		if err != nil {
			t.Fatal(err)
		}
		entity := Entity{
			Id:        1,
			Name:      "John",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.On("BeginTransaction").Return(tx, nil)
		repo.On("FindByNameTx", tx, entity.Name).Return(true, nil)
		id, err := srv.Add(NameRequest{entity.Name})
		assert.Error(t, err)
		assert.Equal(t, int64(0), id)
		a.True(repo.AssertNumberOfCalls(t, "FindByNameTx", 1))
		a.True(repo.AssertNumberOfCalls(t, "BeginTransaction", 1))
	})
	t.Run("employee not exists but fall while add", func(t *testing.T) {
		a := assert.New(t)
		db, mockDB, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := db.Close()
			if err != nil {
				t.Fatal("error close database")
			}
		}()
		sqlxDB := sqlx.NewDb(db, "sqlmock")
		mockDB.ExpectBegin()
		mockDB.ExpectRollback()
		mockDB.ExpectClose()
		tx, err := sqlxDB.Beginx()
		if err != nil {
			t.Fatal(err)
		}
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		if err != nil {
			t.Fatal(err)
		}
		entity := Entity{
			Id:        1,
			Name:      "John",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		want := fmt.Errorf("employee service: add employee: error adding employee")
		repo.On("BeginTransaction").Return(tx, nil)
		repo.On("FindByNameTx", mock.Anything, entity.Name).Return(false, nil)
		repo.On("Add", mock.Anything, mock.MatchedBy(func(e Entity) bool {
			return e.Name == "John"
		})).Return(int64(-1), want)
		id, err := srv.Add(NameRequest{entity.Name})
		a.Error(err)
		a.Contains(err.Error(), want.Error())
		a.Equal(int64(-1), id)
		a.True(repo.AssertNumberOfCalls(t, "Add", 1))
	})
	t.Run("should return added employee's id", func(t *testing.T) {
		db, mockDB, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := db.Close()
			if err != nil {
				t.Fatal("error close database")
			}
		}()
		sqlxDB := sqlx.NewDb(db, "sqlmock")
		mockDB.ExpectBegin()
		mockDB.ExpectCommit()
		mockDB.ExpectClose()
		tx, err := sqlxDB.Beginx()
		if err != nil {
			t.Fatal(err)
		}
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		if err != nil {
			t.Fatal(err)
		}
		entity := Entity{
			Id: 1, Name: "John", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		want := int64(1)
		repo.On("BeginTransaction").Return(tx, nil)
		repo.On("FindByNameTx", mock.Anything, entity.Name).Return(false, nil)
		repo.On("Add", mock.Anything, mock.MatchedBy(func(e Entity) bool {
			return e.Name == "John"
		})).Return(want, nil)
		got, err := srv.Add(NameRequest{entity.Name})
		a.NoError(err)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "Add", 1))
	})
}
func TestGetAll(t *testing.T) {
	a := assert.New(t)
	t.Run("should return employees", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		entity1 := Entity{
			Id: 1, Name: "John", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		entity2 := Entity{
			Id: 2, Name: "Ivan", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		entity3 := Entity{
			Id: 3, Name: "Mr. Smith", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		entities := []Entity{entity1, entity2, entity3}
		want := []Response{}
		for _, e := range entities {
			want = append(want, e.toResponse())
		}
		repo.On("GetAll").Return(entities, nil)
		got, err := srv.GetAll(context.Background())
		a.NoError(err)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "GetAll", 1))
	})
	t.Run("should return wrapped error", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		entities := []Entity{}
		want := fmt.Errorf("employee service: get all employees: error to retrieve all employees")
		repo.On("GetAll").Return(entities, want)
		response, got := srv.GetAll(context.Background())
		a.Empty(response)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "GetAll", 1))
	})
}

func TestGetGroupById(t *testing.T) {
	a := assert.New(t)
	t.Run("should return employees by ids", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		entity1 := Entity{
			Id: 1, Name: "John", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		entity2 := Entity{
			Id: 2, Name: "Ivan", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		entity3 := Entity{
			Id: 3, Name: "Mr. Smith", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		entities := []Entity{entity1, entity2, entity3}
		want := []Response{}
		for _, e := range entities {
			want = append(want, e.toResponse())
		}
		ids := []int64{1, 2}
		repo.On("GetGroupById", ids).Return(entities, nil)
		got, err := srv.GetGroupById(IdsRequest{ids})
		a.NoError(err)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "GetGroupById", 1))
	})
	t.Run("should return wrapped error", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		entities := []Entity{}
		err := errors.New("database error")
		ids := []int64{1, 2}
		want := fmt.Errorf("employee service: get group by id: error getting employees with ids %v", ids)
		repo.On("GetGroupById", ids).Return(entities, err)
		response, got := srv.GetGroupById(IdsRequest{ids})
		a.Empty(response)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "GetGroupById", 1))
	})
}
func TestDelete(t *testing.T) {
	a := assert.New(t)
	t.Run("should delete employee by id", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		repo.On("Delete", int64(1)).Return(nil)
		err := srv.Delete(IdRequest{int64(1)})
		a.NoError(err)
		a.True(repo.AssertNumberOfCalls(t, "Delete", 1))
	})
	t.Run("should return wrapped error", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		err := errors.New("database error")
		id := int64(1)
		want := fmt.Errorf("employee service: delete: error deleting employee with id %d", id)
		repo.On("Delete", id).Return(err)
		got := srv.Delete(IdRequest{id})
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "Delete", 1))
	})
}
func TestDeleteGroup(t *testing.T) {
	a := assert.New(t)
	t.Run("should delete employees by ids", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		ids := []int64{1, 2}
		repo.On("DeleteGroup", ids).Return(nil)
		err := srv.DeleteGroup(IdsRequest{ids})
		a.NoError(err)
		a.True(repo.AssertNumberOfCalls(t, "DeleteGroup", 1))
	})
	t.Run("should return wrapped error", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		err := errors.New("database error")
		ids := []int64{1, 2}
		want := fmt.Errorf("employee service: delete group: error deleting group with id %v", ids)
		repo.On("DeleteGroup", ids).Return(err)
		got := srv.DeleteGroup(IdsRequest{ids})
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "DeleteGroup", 1))
	})
}

func TestValidator_IdRequest(t *testing.T) {
	a := assert.New(t)
	v := validator.New()
	tests := []struct {
		name      string
		input     IdRequest
		wantError bool
		errorHint string
	}{
		{name: "correct id", input: IdRequest{1}, wantError: false, errorHint: ""},
		{name: "zero id", input: IdRequest{0}, wantError: true, errorHint: "Field validation for 'Id' failed on the 'gt' tag"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.input)
			if tt.wantError {
				a.Error(err)
				a.Contains(err.Error(), tt.errorHint)
			} else {
				a.NoError(err)
			}
		})
	}
}

func TestValidator_IdsRequest(t *testing.T) {
	a := assert.New(t)
	v := validator.New()
	tests := []struct {
		name      string
		input     IdsRequest
		wantError bool
		errorHint string
	}{
		{name: "correct ids", input: IdsRequest{[]int64{1, 2, 3}}, wantError: false, errorHint: ""},
		{name: "short ids", input: IdsRequest{}, wantError: true, errorHint: "Field validation for 'Ids' failed on the 'required' tag"},
		{name: "zero id", input: IdsRequest{[]int64{1, 2, 0, 3}}, wantError: true, errorHint: "Field validation for 'Ids[2]' failed on the 'gt' tag"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.input)
			if tt.wantError {
				a.Error(err)
				a.Contains(err.Error(), tt.errorHint)
			} else {
				a.NoError(err)
			}
		})
	}
}

func TestValidator_NameRequest(t *testing.T) {
	a := assert.New(t)
	v := validator.New()
	tests := []struct {
		name      string
		input     NameRequest
		wantError bool
		errorHint string
	}{
		{name: "correct name", input: NameRequest{"Ivan"}, wantError: false, errorHint: ""},
		{name: "empty name", input: NameRequest{""}, wantError: true, errorHint: ""},
		{name: "short name", input: NameRequest{"a"}, wantError: true, errorHint: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.input)
			if tt.wantError {
				a.Error(err)
				a.Contains(err.Error(), tt.errorHint)
			} else {
				a.NoError(err)
			}
		})
	}
}
func TestPageRequestValidation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		input   PageRequest
		wantErr bool
	}{
		{
			name: "valid input",
			input: PageRequest{
				PageNumber: 0,
				PageSize:   10,
			},
			wantErr: false,
		},
		{
			name: "PageSize < 1",
			input: PageRequest{
				PageNumber: 1,
				PageSize:   0,
			},
			wantErr: true,
		},
		{
			name: "PageSize > 100",
			input: PageRequest{
				PageNumber: 2,
				PageSize:   101,
			},
			wantErr: true,
		},
		{
			name: "PageNumber < 0",
			input: PageRequest{
				PageNumber: -1,
				PageSize:   20,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
