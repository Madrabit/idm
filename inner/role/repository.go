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
	CreateAt time.Time `db:"create_at"`
	UpdateAt time.Time `db:"update_at"`
}

func (r *RoleRepository) FindById(id int64) (role RoleEntity, err error) {
	err = r.db.Get(&role, "SELECT * FROM role WHERE id=?", id)
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
	_, err := r.FindById(role.Id)
	if err == nil {
		return -1, fmt.Errorf("Role %d already exists ", role.Id)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return -1, fmt.Errorf("db error %w", err)
	}
	_, err = r.db.Exec("INSERT INTO Role (Id, Name) VALUES (?, ?)", role.Id, role.Name)
	if err != nil {
		return -1, err
	}
	return role.Id, nil
}

func (r *RoleRepository) GetGroupById(ids []int64) (roles []RoleEntity, err error) {
	if len(ids) == 0 {
		return nil, errors.New("Role id can not be empty")
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

func (r *RoleRepository) Delete(role RoleEntity) (err error) {
	e, err := r.FindById(role.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("role do not exists %d", role.Id)
		}
		return fmt.Errorf("db error %w", err)
	}
	_, err = r.db.Exec("DELETE FROM role WHERE id=?", e.Id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %v", err)
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
