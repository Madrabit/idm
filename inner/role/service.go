package role

import (
	"database/sql"
	"errors"
	"fmt"
	"idm/inner/common"
)

type Service struct {
	repo      Repo
	validator Validator
}

type Repo interface {
	FindById(id int64) (Entity, error)
	GetAll() ([]Entity, error)
	Add(role Entity) (int64, error)
	GetGroupById(ids []int64) ([]Entity, error)
	Delete(id int64) error
	DeleteGroup(ids []int64) error
}

type Validator interface {
	Validate(request any) error
}

func NewService(repo Repo, validator Validator) *Service {
	return &Service{repo: repo, validator: validator}
}

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func (s *Service) FindById(req IdRequest) (role Response, err error) {
	if err = s.validator.Validate(req); err != nil {
		return Response{}, &common.RequestValidationError{Massage: err.Error()}
	}
	entity, err := s.repo.FindById(req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Response{}, &common.NotFoundError{Massage: fmt.Sprintf("service repository: find by id: "+
				"role not found: id=%d", req.Id)}
		}
		return Response{}, fmt.Errorf("service repository: find by id: error finding role: id=%d", req.Id)
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

func (s *Service) Add(request NameRequest) (id int64, err error) {
	if err = s.validator.Validate(request); err != nil {
		return 0, &common.RequestValidationError{Massage: err.Error()}
	}
	id, err = s.repo.Add(request.toEntity())
	if err != nil {
		return -1, fmt.Errorf("role service: add employee: error adding role")
	}
	return id, nil
}

func (s *Service) GetGroupById(req IdsRequest) ([]Response, error) {
	if err := s.validator.Validate(req); err != nil {
		return []Response{}, &common.RequestValidationError{Massage: err.Error()}
	}
	roles, err := s.repo.GetGroupById(req.Ids)
	if err != nil {
		return nil, fmt.Errorf("role service: get group by id: error getting roles with ids %v", req.Ids)
	}
	var resp []Response
	for _, role := range roles {
		resp = append(resp, role.toResponse())
	}
	return resp, nil
}

func (s *Service) Delete(req IdRequest) error {
	if err := s.validator.Validate(req); err != nil {
		return &common.RequestValidationError{Massage: err.Error()}
	}
	err := s.repo.Delete(req.Id)
	if err != nil {
		return fmt.Errorf("role service: delete: error deleting role with id %d", req.Id)
	}
	return nil
}

func (s *Service) DeleteGroup(req IdsRequest) error {
	if err := s.validator.Validate(req); err != nil {
		return &common.RequestValidationError{Massage: err.Error()}
	}
	err := s.repo.DeleteGroup(req.Ids)
	if err != nil {
		return fmt.Errorf("role service: delete group: error deleting group with id %v", req.Ids)
	}
	return nil
}
