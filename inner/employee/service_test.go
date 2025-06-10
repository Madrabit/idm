package employee

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) FindById(id int64) (Entity, error) {
	args := m.Called(id)
	return args.Get(0).(Entity), args.Error(1)
}

func (m *MockRepo) GetAll() ([]Entity, error) {
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
		srv := NewService(repo)
		entity := Entity{
			Id:       1,
			Name:     "John",
			CreateAt: time.Now(),
			UpdateAt: time.Now(),
		}
		want := entity.toResponse()
		repo.On("FindById", int64(1)).Return(entity, nil)
		got, err := srv.FindById(1)
		a.NoError(err)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "FindById", 1))
	})
	t.Run("should return wrapped error", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo)
		entity := Entity{}
		id := int64(1)
		want := &NotFoundError{fmt.Sprintf("employee service: find by id: employee not found: id=%d", id)}
		repo.On("FindById", id).Return(entity, sql.ErrNoRows)
		response, got := srv.FindById(id)
		var notFoundErr *NotFoundError
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
		srv := NewService(repo)
		if err != nil {
			t.Fatal(err)
		}
		entity := Entity{
			Id:       1,
			Name:     "John",
			CreateAt: time.Now(),
			UpdateAt: time.Now(),
		}
		repo.On("BeginTransaction").Return(tx, nil)
		want := fmt.Errorf("rollback failed: original error: employee service: add employee: error checking exists employee")
		repo.On("FindByNameTx", tx, entity.Name).Return(false, want)
		response, got := srv.Add(entity)
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
		srv := NewService(repo)
		if err != nil {
			t.Fatal(err)
		}
		entity := Entity{
			Id:       1,
			Name:     "John",
			CreateAt: time.Now(),
			UpdateAt: time.Now(),
		}
		repo.On("BeginTransaction").Return(tx, nil)
		repo.On("FindByNameTx", tx, entity.Name).Return(true, nil)
		id, err := srv.Add(entity)
		assert.Error(t, err)
		assert.Equal(t, entity.Id, id)
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
		srv := NewService(repo)
		if err != nil {
			t.Fatal(err)
		}
		entity := Entity{
			Id:       1,
			Name:     "John",
			CreateAt: time.Now(),
			UpdateAt: time.Now(),
		}
		want := fmt.Errorf("employee service: add employee: error adding employee")
		repo.On("BeginTransaction").Return(tx, nil)
		repo.On("FindByNameTx", mock.Anything, entity.Name).Return(false, nil)
		repo.On("Add", mock.Anything, entity).Return(int64(-1), want)
		id, err := srv.Add(entity)
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
		srv := NewService(repo)
		if err != nil {
			t.Fatal(err)
		}
		entity := Entity{
			Id: 1, Name: "John", CreateAt: time.Now(), UpdateAt: time.Now(),
		}
		want := int64(1)
		repo.On("BeginTransaction").Return(tx, nil)
		repo.On("FindByNameTx", mock.Anything, entity.Name).Return(false, nil)
		repo.On("Add", mock.Anything, entity).Return(want, nil)
		got, err := srv.Add(entity)
		a.NoError(err)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "Add", 1))
	})
}
func TestGetAll(t *testing.T) {
	a := assert.New(t)
	t.Run("should return employees", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo)
		entity1 := Entity{
			Id: 1, Name: "John", CreateAt: time.Now(), UpdateAt: time.Now(),
		}
		entity2 := Entity{
			Id: 2, Name: "Ivan", CreateAt: time.Now(), UpdateAt: time.Now(),
		}
		entity3 := Entity{
			Id: 3, Name: "Mr. Smith", CreateAt: time.Now(), UpdateAt: time.Now(),
		}
		entities := []Entity{entity1, entity2, entity3}
		want := []Response{}
		for _, e := range entities {
			want = append(want, e.toResponse())
		}
		repo.On("GetAll").Return(entities, nil)
		got, err := srv.GetAll()
		a.NoError(err)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "GetAll", 1))
	})
	t.Run("should return wrapped error", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo)
		entities := []Entity{}
		want := fmt.Errorf("employee service: get all employees: error to retrieve all employees")
		repo.On("GetAll").Return(entities, want)
		response, got := srv.GetAll()
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
		srv := NewService(repo)
		entity1 := Entity{
			Id: 1, Name: "John", CreateAt: time.Now(), UpdateAt: time.Now(),
		}
		entity2 := Entity{
			Id: 2, Name: "Ivan", CreateAt: time.Now(), UpdateAt: time.Now(),
		}
		entity3 := Entity{
			Id: 3, Name: "Mr. Smith", CreateAt: time.Now(), UpdateAt: time.Now(),
		}
		entities := []Entity{entity1, entity2, entity3}
		want := []Response{}
		for _, e := range entities {
			want = append(want, e.toResponse())
		}
		ids := []int64{1, 2}
		repo.On("GetGroupById", ids).Return(entities, nil)
		got, err := srv.GetGroupById(ids)
		a.NoError(err)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "GetGroupById", 1))
	})
	t.Run("should return wrapped error", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo)
		entities := []Entity{}
		err := errors.New("database error")
		ids := []int64{1, 2}
		want := fmt.Errorf("employee service: get group by id: error getting employees with ids %v", ids)
		repo.On("GetGroupById", ids).Return(entities, err)
		response, got := srv.GetGroupById(ids)
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
		srv := NewService(repo)
		repo.On("Delete", int64(1)).Return(nil)
		err := srv.Delete(int64(1))
		a.NoError(err)
		a.True(repo.AssertNumberOfCalls(t, "Delete", 1))
	})
	t.Run("should return wrapped error", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo)
		err := errors.New("database error")
		id := int64(1)
		want := fmt.Errorf("employee service: delete: error deleting employee with id %d", id)
		repo.On("Delete", id).Return(err)
		got := srv.Delete(id)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "Delete", 1))
	})
}
func TestDeleteGroup(t *testing.T) {
	a := assert.New(t)
	t.Run("should delete employees by ids", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo)
		ids := []int64{1, 2}
		repo.On("DeleteGroup", ids).Return(nil)
		err := srv.DeleteGroup(ids)
		a.NoError(err)
		a.True(repo.AssertNumberOfCalls(t, "DeleteGroup", 1))
	})
	t.Run("should return wrapped error", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo)
		err := errors.New("database error")
		ids := []int64{1, 2}
		want := fmt.Errorf("employee service: delete group: error deleting group with id %v", ids)
		repo.On("DeleteGroup", ids).Return(err)
		got := srv.DeleteGroup(ids)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "DeleteGroup", 1))
	})
}
