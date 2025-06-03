package employee

import "fmt"

type Service struct {
	repo Repo
}

type Repo interface {
	FindById(id int64) (Entity, error)
	GetAll() ([]Entity, error)
	Add(employee Entity) (int64, error)
	GetGroupById(ids []int64) ([]Entity, error)
	Delete(id int64) error
	DeleteGroup(ids []int64) error
}

func NewService(repo Repo) *Service {
	return &Service{repo}
}

func (s *Service) FindById(id int64) (role Response, err error) {
	entity, err := s.repo.FindById(id)
	if err != nil {
		return Response{}, fmt.Errorf("error finding employee with id %d: %w", id, err)
	}
	return entity.toResponse(), nil
}

func (s *Service) GetAll() ([]Response, error) {
	all, err := s.repo.GetAll()
	if err != nil {
		return []Response{}, fmt.Errorf("error getting all employees: %w", err)
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
		return -1, fmt.Errorf("error adding employee: %w", err)
	}
	return id, nil
}

func (s *Service) GetGroupById(ids []int64) ([]Response, error) {
	roles, err := s.repo.GetGroupById(ids)
	if err != nil {
		return nil, fmt.Errorf("error getting employee with id %v: %w", ids, err)
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
		return fmt.Errorf("error deleting employee with id %d: %w", id, err)
	}
	return nil
}

func (s *Service) DeleteGroup(ids []int64) error {
	err := s.repo.DeleteGroup(ids)
	if err != nil {
		return fmt.Errorf("error deleting group with id %v: %w", ids, err)
	}
	return nil
}
