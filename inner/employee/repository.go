package employee

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) BeginTransaction() (tx *sqlx.Tx, err error) {
	return r.db.Beginx()
}

func (r *Repository) FindById(id int64) (employee Entity, err error) {
	err = r.db.Get(&employee, "SELECT * FROM employee WHERE id=$1", id)
	return employee, err
}

func (r *Repository) FindByNameTx(tx *sqlx.Tx, name string) (isExists bool, err error) {
	err = tx.Get(
		&isExists,
		"select exists(select 1 from employee where name = $1)",
		name,
	)
	return isExists, err
}

func (r *Repository) GetAll(ctx context.Context) ([]Entity, error) {
	var employees []Entity
	rows, err := r.db.QueryxContext(ctx, "SELECT * FROM employee")
	if err != nil {
		return employees, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			return
		}
	}()
	for rows.Next() {
		var employee Entity
		if err = rows.StructScan(&employee); err != nil {
			return nil, err
		}
		employees = append(employees, employee)
	}
	return employees, nil
}

func (r *Repository) Add(tx *sqlx.Tx, employee Entity) (int64, error) {
	var id int64
	err := tx.QueryRow("INSERT INTO employee (name) VALUES ($1) RETURNING id",
		employee.Name).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (r *Repository) GetGroupById(ids []int64) (employees []Entity, err error) {
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
	_, err = r.db.Exec("DELETE FROM employee WHERE id=$1", id)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) DeleteGroup(ids []int64) error {
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
