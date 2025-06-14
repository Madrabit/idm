package employee

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	repo Repo
}

type Repo interface {
	FindById(id int64) (Entity, error)
	FindByNameTx(tx *sqlx.Tx, name string) (bool, error)
	GetAll() ([]Entity, error)
	Add(tx *sqlx.Tx, employee Entity) (int64, error)
	GetGroupById(ids []int64) ([]Entity, error)
	Delete(id int64) error
	DeleteGroup(ids []int64) error
	BeginTransaction() (*sqlx.Tx, error)
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
			return Response{}, &NotFoundError{fmt.Sprintf("employee service: find by id: "+
				"employee not found: id=%d", id)}
		}
		return Response{}, fmt.Errorf("employee service: find by id: error finding employee: id=%d", id)
	}
	return entity.toResponse(), nil
}

func (s *Service) GetAll() ([]Response, error) {
	all, err := s.repo.GetAll()
	if err != nil {
		return []Response{}, fmt.Errorf("employee service: get all employees: error to retrieve all employees")
	}
	var resp []Response
	for _, entity := range all {
		resp = append(resp, entity.toResponse())
	}
	return resp, nil
}

func (s *Service) Add(employee Entity) (id int64, err error) {
	tx, err := s.repo.BeginTransaction()
	if err != nil {
		return 0, fmt.Errorf("employee service: add employee: error starting transaction")
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			err = fmt.Errorf("employee service: add employee: panic add employee: %v", p)
			return
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("rollback failed: original error: %w", err)
			}
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			err = fmt.Errorf("employee service: add employee: committing transaction failed: %w", commitErr)
		}
	}()
	isExists, err := s.repo.FindByNameTx(tx, employee.Name)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("employee service: add employee: error checking exists employee")
	}
	if isExists {
		return employee.Id, nil
	}
	id, err = s.repo.Add(tx, employee)
	if err != nil {
		return -1, fmt.Errorf("employee service: add employee: error adding employee")
	}
	return id, nil
}

func (s *Service) GetGroupById(ids []int64) ([]Response, error) {
	employees, err := s.repo.GetGroupById(ids)
	if err != nil {
		return nil, fmt.Errorf("employee service: get group by id: error getting employees with ids %v", ids)
	}
	var resp []Response
	for _, emp := range employees {
		resp = append(resp, emp.toResponse())
	}
	return resp, nil
}

func (s *Service) Delete(id int64) error {
	err := s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("employee service: delete: error deleting employee with id %d", id)
	}
	return nil
}

func (s *Service) DeleteGroup(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	err := s.repo.DeleteGroup(ids)
	if err != nil {
		return fmt.Errorf("employee service: delete group: error deleting group with id %v", ids)
	}
	return nil
}
