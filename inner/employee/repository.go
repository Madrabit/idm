package employee

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

type EmployeeRepository struct {
	db *sqlx.DB
}

func NewEmployeeRepository(db *sqlx.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

type EmployeeEntity struct {
	Id       int64     `db:"id"`
	Name     string    `db:"name"`
	CreateAt time.Time `db:"create_at"`
	UpdateAt time.Time `db:"update_at"`
}

func (r *EmployeeRepository) FindById(id int64) (employee EmployeeEntity, err error) {
	err = r.db.Get(&employee, "SELECT * FROM employee WHERE id=?", id)
	return employee, err
}

func (r *EmployeeRepository) GetAll() ([]EmployeeEntity, error) {
	var employees []EmployeeEntity
	rows, err := r.db.Queryx("SELECT * FROM employee")
	if err != nil {
		return employees, err
	}
	defer rows.Close()
	for rows.Next() {
		var employee EmployeeEntity
		if err = rows.StructScan(&employee); err != nil {
			return nil, err
		}
		employees = append(employees, employee)
	}
	return employees, nil
}

func (r *EmployeeRepository) Add(employee EmployeeEntity) (int64, error) {
	_, err := r.FindById(employee.Id)
	if err == nil {
		return -1, fmt.Errorf("employee %d already exists ", employee.Id)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return -1, fmt.Errorf("db error %w", err)
	}
	_, err = r.db.Exec("INSERT INTO employee (Id, Name) VALUES (?, ?)", employee.Id, employee.Name)
	if err != nil {
		return -1, err
	}
	return employee.Id, nil
}

func (r *EmployeeRepository) GetGroupById(ids []int64) (employees []EmployeeEntity, err error) {
	if len(ids) == 0 {
		return nil, errors.New("employee id can not be empty")
	}
	q, args, err := sqlx.In("SELECT * FROM employee WHERE id IN (?)", ids)
	if err != nil {
		return nil, err
	}
	q = r.db.Rebind(q)
	err = r.db.Select(&employees, q, args...)
	if err != nil {
		return nil, err
	}
	return employees, nil
}

func (r *EmployeeRepository) Delete(employee EmployeeEntity) (err error) {
	e, err := r.FindById(employee.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("employee do not exists %d", employee.Id)
		}
		return fmt.Errorf("db error %w", err)
	}
	_, err = r.db.Exec("DELETE FROM employee WHERE id=?", e.Id)
	if err != nil {
		return fmt.Errorf("failed to delete employee: %v", err)
	}
	return nil
}

func (r *EmployeeRepository) DeleteGroup(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	q, args, err := sqlx.In("DELETE FROM employee WHERE id IN (?)", ids)
	if err != nil {
		return err
	}
	q = r.db.Rebind(q)
	_, err = r.db.Exec(q, args...)
	if err != nil {
		return err
	}
	return nil
}
