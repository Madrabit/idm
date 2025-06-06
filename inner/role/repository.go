package Role

import (
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindById(id int64) (role Entity, err error) {
	err = r.db.Get(&role, "SELECT * FROM role WHERE id=$1", id)
	return role, err
}

func (r *Repository) GetAll() ([]Entity, error) {
	var roles []Entity
	rows, err := r.db.Queryx("SELECT * FROM role")
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			return
		}
	}()
	for rows.Next() {
		var role Entity
		if err = rows.StructScan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *Repository) Add(role Entity) (int64, error) {
	var id int64
	err := r.db.QueryRow("INSERT INTO Role (name) VALUES ($1) RETURNING id",
		role.Name).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (r *Repository) GetGroupById(ids []int64) (roles []Entity, err error) {
	q, args, err := sqlx.In("SELECT * FROM role WHERE id IN (?)", ids)
	if err != nil {
		return nil, err
	}
	q = r.db.Rebind(q)
	err = r.db.Select(&roles, q, args...)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *Repository) Delete(id int64) (err error) {
	_, err = r.db.Exec("DELETE FROM role WHERE id=$1", id)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) DeleteGroup(ids []int64) error {
	q, args, err := sqlx.In("DELETE FROM role WHERE id IN (?)", ids)
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
