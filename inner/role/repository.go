package Role

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(db *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

type RoleEntity struct {
	Id       int64     `db:"id"`
	Name     string    `db:"name"`
	CreateAt time.Time `db:"created_at"`
	UpdateAt time.Time `db:"updated_at"`
}

func (r *RoleRepository) FindById(id int64) (role RoleEntity, err error) {
	err = r.db.Get(&role, "SELECT * FROM role WHERE id=$1", id)
	return role, err
}

func (r *RoleRepository) GetAll() ([]RoleEntity, error) {
	var roles []RoleEntity
	rows, err := r.db.Queryx("SELECT * FROM role")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var role RoleEntity
		if err = rows.StructScan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *RoleRepository) Add(role RoleEntity) (int64, error) {
	var id int64
	err := r.db.QueryRow("INSERT INTO Role (name) VALUES ($1) RETURNING id",
		role.Name).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (r *RoleRepository) GetGroupById(ids []int64) (roles []RoleEntity, err error) {
	if len(ids) == 0 {
		return nil, errors.New("role id can not be empty")
	}
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

func (r *RoleRepository) Delete(id int64) (err error) {
	e, err := r.FindById(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("role not found: %d", id)
		}
		return fmt.Errorf("db error %w", err)
	}
	_, err = r.db.Exec("DELETE FROM role WHERE id=$1", e.Id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

func (r *RoleRepository) DeleteGroup(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
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
