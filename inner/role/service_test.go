package role

import (
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

func (m *MockRepo) Add(role Entity) (int64, error) {
	args := m.Called(role)
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

type Stub struct {
	Entity
	Err error
}

//goland:noinspection GoUnusedExportedType
type StubRepo interface {
	FindById(id int64) (Entity, error)
	GetAll() ([]Entity, error)
	Add(role Entity) (int64, error)
	GetGroupById(ids []int64) ([]Entity, error)
	Delete(id int64) error
	DeleteGroup(ids []int64) error
}

func _() *Stub {
	return &Stub{}
}

func (s *Stub) FindById(_ int64) (Entity, error) {
	return s.Entity, s.Err
}

func (s *Stub) GetAll() ([]Entity, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Stub) Add(_ Entity) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Stub) GetGroupById(_ []int64) ([]Entity, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Stub) Delete(_ int64) error {
	//TODO implement me
	panic("implement me")
}

func (s *Stub) DeleteGroup(_ []int64) error {
	//TODO implement me
	panic("implement me")
}

func TestFindById(t *testing.T) {
	a := assert.New(t)
	t.Run("should return found role", func(t *testing.T) {
		entity := Entity{
			Id:       1,
			Name:     "Admin",
			CreateAt: time.Now(),
			UpdateAt: time.Now(),
		}
		repo := &Stub{
			entity,
			nil,
		}
		srv := NewService(repo)
		want := entity.toResponse()
		got, err := srv.FindById(1)
		a.NoError(err)
		a.Equal(want, got)
	})
	t.Run("should return found role", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo)
		entity := Entity{
			Id:       1,
			Name:     "Admin",
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
		want := &NotFoundError{fmt.Sprintf("service repository: find by id: "+
			"role not found: id=%d", id)}
		repo.On("FindById", id).Return(entity, want)
		response, got := srv.FindById(id)
		a.Empty(response)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "FindById", 1))
	})
}

func TestAdd(t *testing.T) {
	a := assert.New(t)
	t.Run("should return added role's id", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo)
		entity := Entity{
			Id: 1, Name: "Admin", CreateAt: time.Now(), UpdateAt: time.Now(),
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
		want := &AddError{"role service: add employee: error adding role"}
		repo.On("Add", entity).Return(int64(-1), want)
		_, got := srv.Add(entity)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "Add", 1))
	})
}
func TestGetAll(t *testing.T) {
	a := assert.New(t)
	t.Run("should return roles", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo)
		entity1 := Entity{
			Id: 1, Name: "Admin", CreateAt: time.Now(), UpdateAt: time.Now(),
		}
		entity2 := Entity{
			Id: 2, Name: "Manager", CreateAt: time.Now(), UpdateAt: time.Now(),
		}
		entity3 := Entity{
			Id: 3, Name: "Support", CreateAt: time.Now(), UpdateAt: time.Now(),
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
		want := &RetrieveError{Message: "role service: get all roles: error to retrieve all roles"}
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
	t.Run("should return roles by ids", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo)
		entity1 := Entity{
			Id: 1, Name: "Admin", CreateAt: time.Now(), UpdateAt: time.Now(),
		}
		entity2 := Entity{
			Id: 2, Name: "Manager", CreateAt: time.Now(), UpdateAt: time.Now(),
		}
		entity3 := Entity{
			Id: 3, Name: "Support", CreateAt: time.Now(), UpdateAt: time.Now(),
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
		ids := []int64{1, 2}
		want := &RetrieveError{fmt.Sprintf("role service: get group by id: error getting roles with ids %v", ids)}
		repo.On("GetGroupById", ids).Return(entities, want)
		response, got := srv.GetGroupById(ids)
		a.Empty(response)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "GetGroupById", 1))
	})
}
func TestDelete(t *testing.T) {
	a := assert.New(t)
	t.Run("should delete role by id", func(t *testing.T) {
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
		id := int64(1)
		want := &DeleteError{fmt.Sprintf("role service: delete: error deleting role with id %d", id)}
		repo.On("Delete", id).Return(want)
		got := srv.Delete(id)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "Delete", 1))
	})
}
func TestDeleteGroup(t *testing.T) {
	a := assert.New(t)
	t.Run("should delete roles by ids", func(t *testing.T) {
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
		ids := []int64{1, 2}
		want := &DeleteError{fmt.Sprintf("role service: delete group: error deleting group with id %v", ids)}
		repo.On("DeleteGroup", ids).Return(want)
		got := srv.DeleteGroup(ids)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "DeleteGroup", 1))
	})
}
