package tests

import (
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"idm/inner/employee"
	"log"
	"testing"
	"time"
)

func TestEmployeeRepository(t *testing.T) {
	a := assert.New(t)
	fx := NewFixture()
	defer fx.Close()
	t.Run("find employee by id", func(t *testing.T) {
		repo := fx.employees
		fx.ClearTable()
		newEmpId := mustEmployee(t, fx, "Test name")
		got, err := repo.FindById(newEmpId)
		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got.Id)
		a.NotEmpty(got.Name)
		a.NotEmpty(got.CreateAt)
		a.NotEmpty(got.UpdateAt)
		a.Equal("Test name", got.Name)
	})
	t.Run("get all employees", func(t *testing.T) {
		repo := fx.employees
		fx.ClearTable()
		mustEmployee(t, fx, "name 1")
		mustEmployee(t, fx, "name 2")
		mustEmployee(t, fx, "name 3")
		got, err := repo.GetAll()
		a.Nil(err)
		a.NotEmpty(got)
		a.Len(got, 3)
		for _, v := range got {
			a.NotEmpty(v.Id)
			a.NotEmpty(v.Name)
			a.NotEmpty(v.CreateAt)
			a.NotEmpty(v.UpdateAt)
		}
	})
	t.Run("get group employees by ids", func(t *testing.T) {
		repo := fx.employees
		fx.ClearTable()
		mustEmployee(t, fx, "name 1")
		id2 := mustEmployee(t, fx, "name 2")
		id3 := mustEmployee(t, fx, "name 3")
		id4 := mustEmployee(t, fx, "name 4")
		mustEmployee(t, fx, "name 5")
		got, err := repo.GetGroupById([]int64{id2, id3, id4})
		a.Nil(err)
		a.NotEmpty(got)
		a.Len(got, 3)
		for _, v := range got {
			a.NotEmpty(v.Id)
			a.NotEmpty(v.Name)
			a.NotEmpty(v.CreateAt)
			a.NotEmpty(v.UpdateAt)
		}
	})
	t.Run("delete employee", func(t *testing.T) {
		repo := fx.employees
		fx.ClearTable()
		id := mustEmployee(t, fx, "name 1")
		err := repo.Delete(id)
		a.Nil(err)
		got, err := repo.FindById(id)
		a.NotNil(err)
		a.Empty(got)
	})
	t.Run("delete group of employees", func(t *testing.T) {
		repo := fx.employees
		fx.ClearTable()
		mustEmployee(t, fx, "name 1")
		id2 := mustEmployee(t, fx, "name 2")
		id3 := mustEmployee(t, fx, "name 3")
		id4 := mustEmployee(t, fx, "name 4")
		mustEmployee(t, fx, "name 5")
		ids := []int64{id2, id3, id4}
		err := repo.DeleteGroup(ids)
		a.Nil(err)
		got, err := repo.GetGroupById(ids)
		a.NoError(err)
		a.Len(got, 0)
	})
	t.Run("find employee and insert in one tx", func(t *testing.T) {
		repo := fx.employees
		fx.ClearTable()
		tx, err := repo.BeginTransaction()
		a.NoError(err)
		defer func(tx *sqlx.Tx) {
			err := tx.Rollback()
			if err != nil {
				log.Fatal("rollback error")
			}
		}(tx)
		entity := employee.Entity{
			Id:       1,
			Name:     "name1",
			CreateAt: time.Now(),
			UpdateAt: time.Now(),
		}
		isExist, err := repo.FindByNameTx(tx, entity.Name)
		a.NoError(err)
		a.False(isExist, "should be now employee before add")
		got, err := repo.Add(tx, entity)
		a.NoError(err)
		a.NoError(err)
		a.NotEmpty(got)
		found, err := repo.FindByNameTx(tx, entity.Name)
		a.NoError(err)
		a.True(found)
	})
}
func mustEmployee(t *testing.T, f *Fixture, name string) int64 {
	t.Helper()
	id, err := f.Employee(name)
	require.NoError(t, err)
	return id
}
