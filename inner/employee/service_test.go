package employee

import (
	"errors"
	"fmt"
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

func (m *MockRepo) Add(employee Entity) (int64, error) {
	args := m.Called(employee)
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
		err := errors.New("database error")
		want := fmt.Errorf("error finding employee with id 1: %w", err)
		repo.On("FindById", int64(1)).Return(entity, err)
		response, got := srv.FindById(1)
		a.Empty(response)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "FindById", 1))
	})
}

func TestAdd(t *testing.T) {
	a := assert.New(t)
	t.Run("should return added employee's id", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo)
		entity := Entity{
			Id: 1, Name: "John", CreateAt: time.Now(), UpdateAt: time.Now(),
		}
		want := int64(1)
		repo.On("Add", entity).Return(want, nil)
		got, err := srv.Add(entity)
		a.NoError(err)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "Add", 1))
	})
	t.Run("should return wrapped error", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo)
		entity := Entity{}
		err := errors.New("database error")
		want := fmt.Errorf("error adding employee: %w", err)
		repo.On("Add", entity).Return(int64(-1), err)
		_, got := srv.Add(entity)
		a.NotNil(got)
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
		err := errors.New("database error")
		want := fmt.Errorf("error getting all employees: %w", err)
		repo.On("GetAll").Return(entities, err)
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
		want := fmt.Errorf("error getting employee with id %d: %w", ids, err)
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
		want := fmt.Errorf("error deleting employee with id %d: %w", id, err)
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
		want := fmt.Errorf("error deleting group with id %d: %w", ids, err)
		repo.On("DeleteGroup", ids).Return(err)
		got := srv.DeleteGroup(ids)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "DeleteGroup", 1))
	})
}
