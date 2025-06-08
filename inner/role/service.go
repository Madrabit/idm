package role

import (
	"database/sql"
	"errors"
	"fmt"
)

type Service struct {
	repo Repo
}

type Repo interface {
	FindById(id int64) (Entity, error)
	GetAll() ([]Entity, error)
	Add(role Entity) (int64, error)
	GetGroupById(ids []int64) ([]Entity, error)
	Delete(id int64) error
	DeleteGroup(ids []int64) error
}

func NewService(repo Repo) *Service {
	return &Service{repo}
}

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func (s *Service) FindById(id int64) (role Response, err error) {
	entity, err := s.repo.FindById(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Response{}, &NotFoundError{fmt.Sprintf("service repository: find by id: "+
				"role not found: id=%d", id)}
		}
		return Response{}, fmt.Errorf("service repository: find by id: error finding role: id=%d", id)
	}
	return entity.toResponse(), nil
}

func (s *Service) GetAll() ([]Response, error) {
	all, err := s.repo.GetAll()
	if err != nil {
		return []Response{}, fmt.Errorf("role service: get all roles: error to retrieve all roles")
	}
	var resp []Response
	for _, entity := range all {
		resp = append(resp, entity.toResponse())
	}
	return resp, nil
}

func (s *Service) Add(role Entity) (int64, error) {
	id, err := s.repo.Add(role)
	if err != nil {
		return -1, fmt.Errorf("role service: add employee: error adding role")
	}
	return id, nil
}

func (s *Service) GetGroupById(ids []int64) ([]Response, error) {
	roles, err := s.repo.GetGroupById(ids)
	if err != nil {
		return nil, fmt.Errorf("role service: get group by id: error getting roles with ids %v", ids)
	}
	var resp []Response
	for _, role := range roles {
		resp = append(resp, role.toResponse())
	}
	return resp, nil
}

func (s *Service) Delete(id int64) error {
	err := s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("role service: delete: error deleting role with id %d", id)
	}
	return nil
}

func (s *Service) DeleteGroup(ids []int64) error {
	err := s.repo.DeleteGroup(ids)
	if err != nil {
		return fmt.Errorf("role service: delete group: error deleting group with id %v", ids)
	}
	return nil
}
