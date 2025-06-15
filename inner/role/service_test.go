package role

import (
	"database/sql"
	"errors"
	"fmt"
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
		srv := NewService(repo, validator.New())
		want := entity.toResponse()
		got, err := srv.FindById(IdRequest{1})
		a.NoError(err)
		a.Equal(want, got)
	})
	t.Run("should return found role", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		entity := Entity{
			Id:       1,
			Name:     "Admin",
			CreateAt: time.Now(),
			UpdateAt: time.Now(),
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
		want := &common.NotFoundError{Massage: fmt.Sprintf("service repository: find by id: "+
			"role not found: id=%d", id)}
		repo.On("FindById", id).Return(entity, sql.ErrNoRows)
		response, got := srv.FindById(IdRequest{1})
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
	t.Run("should return added role's id", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		entity := Entity{
			Id: 1, Name: "Admin", CreateAt: time.Now(), UpdateAt: time.Now(),
		}
		want := int64(1)
		repo.On("Add", mock.MatchedBy(func(e Entity) bool {
			return e.Name == entity.Name
		})).Return(want, nil)
		got, err := srv.Add(NameRequest{entity.Name})
		a.NoError(err)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "Add", 1))
	})
	t.Run("should return wrapped error", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		_, got := srv.Add(NameRequest{}) // Пустое имя
		a.NotNil(got)
		a.IsType(&common.RequestValidationError{}, got)
		a.Contains(got.Error(), "validation")
		a.True(repo.AssertNumberOfCalls(t, "Add", 0))
	})
}
func TestGetAll(t *testing.T) {
	a := assert.New(t)
	t.Run("should return roles", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
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
		srv := NewService(repo, validator.New())
		entities := []Entity{}
		want := fmt.Errorf("role service: get all roles: error to retrieve all roles")
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
		srv := NewService(repo, validator.New())
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
		got, err := srv.GetGroupById(IdsRequest{ids})
		a.NoError(err)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "GetGroupById", 1))
	})
	t.Run("should return wrapped error", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		entities := []Entity{}
		ids := []int64{1, 2}
		want := fmt.Errorf("role service: get group by id: error getting roles with ids %v", ids)
		repo.On("GetGroupById", ids).Return(entities, want)
		response, got := srv.GetGroupById(IdsRequest{ids})
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
		srv := NewService(repo, validator.New())
		repo.On("Delete", int64(1)).Return(nil)
		err := srv.Delete(IdRequest{int64(1)})
		a.NoError(err)
		a.True(repo.AssertNumberOfCalls(t, "Delete", 1))
	})
	t.Run("should return wrapped error", func(t *testing.T) {
		repo := new(MockRepo)
		srv := NewService(repo, validator.New())
		id := int64(1)
		want := fmt.Errorf("role service: delete: error deleting role with id %d", id)
		repo.On("Delete", id).Return(want)
		got := srv.Delete(IdRequest{id})
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "Delete", 1))
	})
}
func TestDeleteGroup(t *testing.T) {
	a := assert.New(t)
	t.Run("should delete roles by ids", func(t *testing.T) {
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
		ids := []int64{1, 2}
		want := fmt.Errorf("role service: delete group: error deleting group with id %v", ids)
		repo.On("DeleteGroup", ids).Return(want)
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
		{name: "short ids", input: IdsRequest{}, wantError: true, errorHint: "Field validation for 'Ids' failed on the 'min' tag"},
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
