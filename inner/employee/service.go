package employee

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"idm/inner/common"
)

type Service struct {
	repo      Repo
	validator Validator
}

type Repo interface {
	FindById(id int64) (Entity, error)
	FindByNameTx(tx *sqlx.Tx, name string) (bool, error)
	GetAll(ctx context.Context) ([]Entity, error)
	Add(tx *sqlx.Tx, employee Entity) (int64, error)
	GetGroupById(ids []int64) ([]Entity, error)
	Delete(id int64) error
	DeleteGroup(ids []int64) error
	BeginTransaction() (*sqlx.Tx, error)
	FindPageWithFilter(tx *sqlx.Tx, offset, limit int64, name string) (employees []Entity, err error)
	GetTotal(tx *sqlx.Tx, name string) (count int64, err error)
	FindKeySetPagination(tx *sqlx.Tx, lastId, limit int64) ([]Entity, error)
}

type Validator interface {
	Validate(request any) error
}

func NewService(repo Repo, validator Validator) *Service {
	return &Service{repo: repo, validator: validator}
}

func (s *Service) FindById(req IdRequest) (employee Response, err error) {
	if err = s.validator.Validate(req); err != nil {
		return Response{}, &common.RequestValidationError{Massage: err.Error()}
	}
	entity, err := s.repo.FindById(req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Response{}, &common.NotFoundError{Massage: fmt.Sprintf("employee service: find by id: "+
				"employee not found: id=%d", req.Id)}
		}
		return Response{}, fmt.Errorf("employee service: find by id: error finding employee: id=%d", req.Id)
	}
	return entity.toResponse(), nil
}

func (s *Service) GetAll(ctx context.Context) ([]Response, error) {
	all, err := s.repo.GetAll(ctx)
	if err != nil {
		return []Response{}, fmt.Errorf("employee service: get all employees: error to retrieve all employees")
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
	isExists, err := s.repo.FindByNameTx(tx, request.Name)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("employee service: add employee: error checking exists employee")
	}
	if isExists {
		return 0, &common.AlreadyExistsError{Massage: fmt.Sprintf("employee with name %s already exists", request.Name)}
	}
	id, err = s.repo.Add(tx, request.toEntity())
	if err != nil {
		return -1, fmt.Errorf("employee service: add employee: error adding employee")
	}
	return id, nil
}

func (s *Service) GetGroupById(req IdsRequest) ([]Response, error) {
	if err := s.validator.Validate(req); err != nil {
		return []Response{}, &common.RequestValidationError{Massage: err.Error()}
	}
	employees, err := s.repo.GetGroupById(req.Ids)
	if err != nil {
		return nil, fmt.Errorf("employee service: get group by id: error getting employees with ids %v", req.Ids)
	}
	var resp []Response
	for _, emp := range employees {
		resp = append(resp, emp.toResponse())
	}
	return resp, nil
}

func (s *Service) Delete(req IdRequest) error {
	if err := s.validator.Validate(req); err != nil {
		return &common.RequestValidationError{Massage: err.Error()}
	}
	err := s.repo.Delete(req.Id)
	if err != nil {
		return fmt.Errorf("employee service: delete: error deleting employee with id %d", req.Id)
	}
	return nil
}

func (s *Service) DeleteGroup(req IdsRequest) error {
	if err := s.validator.Validate(req); err != nil {
		return &common.RequestValidationError{Massage: err.Error()}
	}
	err := s.repo.DeleteGroup(req.Ids)
	if err != nil {
		return fmt.Errorf("employee service: delete group: error deleting group with id %v", req.Ids)
	}
	return nil
}

func (s *Service) GetPage(request PageRequest) (pageEmp PageResponse, err error) {
	if err = s.validator.Validate(request); err != nil {
		return PageResponse{}, &common.RequestValidationError{Massage: err.Error()}
	}
	tx, err := s.repo.BeginTransaction()
	if err != nil {
		return PageResponse{}, fmt.Errorf("employee service: get page: error starting transaction")
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			err = fmt.Errorf("employee service: get page: panic add employee: %v", p)
			return
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("rollback failed: original error: %w", err)
			}
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			err = fmt.Errorf("employee service: get page: committing transaction failed: %w", commitErr)
		}
	}()
	offset := request.PageNumber * request.PageSize
	limit := request.PageSize
	name := request.TextFilter
	var page []Entity
	page, err = s.repo.FindPageWithFilter(tx, offset, limit, name)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return PageResponse{}, fmt.Errorf("employee service: get page")
	}
	total, err := s.repo.GetTotal(tx, name)
	if err != nil {
		return PageResponse{}, fmt.Errorf("employee service: get total count of page")
	}
	pageEmp = PageResponse{
		page,
		request.PageSize,
		request.PageNumber,
		total,
	}
	return pageEmp, nil
}

func (s *Service) GetKeySetPage(request PageKeySetRequest) (pageEmp PageKeySetResponse, err error) {
	if err = s.validator.Validate(request); err != nil {
		return PageKeySetResponse{}, &common.RequestValidationError{Massage: err.Error()}
	}
	tx, err := s.repo.BeginTransaction()
	if err != nil {
		return PageKeySetResponse{}, fmt.Errorf("employee service: get page: error starting transaction")
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			err = fmt.Errorf("employee service: get page: panic add employee: %v", p)
			return
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("rollback failed: original error: %w", err)
			}
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			err = fmt.Errorf("employee service: get page: committing transaction failed: %w", commitErr)
		}
	}()
	lastId := request.LastId
	limit := request.PageSize
	name := request.TextFilter
	page, err := s.repo.FindKeySetPagination(tx, lastId, limit)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return PageKeySetResponse{}, fmt.Errorf("employee service: get page")
	}
	total, err := s.repo.GetTotal(tx, name)
	if err != nil {
		return PageKeySetResponse{}, fmt.Errorf("employee service: get total count of page")
	}
	pageEmp = PageKeySetResponse{
		Result: page,
		LastId: lastId,
		Total:  total,
	}
	return pageEmp, nil
}
