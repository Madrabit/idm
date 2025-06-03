package employee

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewEmployeeRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindById(id int64) (employee Entity, err error) {
	err = r.db.Get(&employee, "SELECT * FROM employee WHERE id=$1", id)
	return employee, err
}

func (r *Repository) GetAll() ([]Entity, error) {
	var employees []Entity
	rows, err := r.db.Queryx("SELECT * FROM employee")
	if err != nil {
		return employees, err
	}
	defer rows.Close()
	for rows.Next() {
		var employee Entity
		if err = rows.StructScan(&employee); err != nil {
			return nil, err
		}
		employees = append(employees, employee)
	}
	return employees, nil
}

func (r *Repository) Add(employee Entity) (int64, error) {
	var id int64
	err := r.db.QueryRow("INSERT INTO employee (name) VALUES ($1) RETURNING id",
		employee.Name).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (r *Repository) GetGroupById(ids []int64) (employees []Entity, err error) {
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

func (r *Repository) Delete(id int64) (err error) {
	e, err := r.FindById(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("employee not found: %d", id)
		}
		return fmt.Errorf("db error %w", err)
	}
	_, err = r.db.Exec("DELETE FROM employee WHERE id=$1", e.Id)
	if err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}
	return nil
}

func (r *Repository) DeleteGroup(ids []int64) error {
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
