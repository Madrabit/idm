package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRoleRepository(t *testing.T) {
	a := assert.New(t)
	fx := NewRoleFixture()
	defer fx.Close()
	t.Run("find role by id", func(t *testing.T) {
		repo := fx.repo
		fx.ClearTable()
		newEmpId := mustRole(t, fx, "Test name")
		got, err := repo.FindById(newEmpId)
		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got.Id)
		a.NotEmpty(got.Name)
		a.NotEmpty(got.CreateAt)
		a.NotEmpty(got.UpdateAt)
		a.Equal("Test name", got.Name)
	})
	t.Run("get all roles", func(t *testing.T) {
		repo := fx.repo
		fx.ClearTable()
		mustRole(t, fx, "name 1")
		mustRole(t, fx, "name 1")
		mustRole(t, fx, "name 1")
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
	t.Run("get group roles by ids", func(t *testing.T) {
		repo := fx.repo
		fx.ClearTable()
		mustRole(t, fx, "name 1")
		id2 := mustRole(t, fx, "name 2")
		id3 := mustRole(t, fx, "name 3")
		id4 := mustRole(t, fx, "name 4")
		mustRole(t, fx, "name 5")
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
	t.Run("delete role", func(t *testing.T) {
		repo := fx.repo
		fx.ClearTable()
		id := mustRole(t, fx, "name 1")
		err := repo.Delete(id)
		a.Nil(err)
		got, err := repo.FindById(id)
		a.NotNil(err)
		a.Empty(got)
	})
	t.Run("delete group of roles", func(t *testing.T) {
		repo := fx.repo
		fx.ClearTable()
		mustRole(t, fx, "name 1")
		id2 := mustRole(t, fx, "name 2")
		id3 := mustRole(t, fx, "name 3")
		id4 := mustRole(t, fx, "name 4")
		mustRole(t, fx, "name 5")
		ids := []int64{id2, id3, id4}
		err := repo.DeleteGroup(ids)
		a.Nil(err)
		got, err := repo.GetGroupById(ids)
		a.NoError(err)
		a.Len(got, 0)
	})
}

func mustRole(t *testing.T, f *RoleFixture, name string) int64 {
	t.Helper()
	id, err := f.Role(name)
	require.NoError(t, err)
	return id
}
