package repository

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"

	"packster/pkg/types"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPermissionRepo(t *testing.T) (*PermissionRepo, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return &PermissionRepo{SqlDB: db}, mock, func() { db.Close() }
}

func TestPermissionRepo_Get_NotFound(t *testing.T) {
	repo, mock, cleanup := newPermissionRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT account, project`).
		WithArgs(1, 2).
		WillReturnError(sql.ErrNoRows)

	got, err := repo.Get(1, 2)
	assert.NoError(t, err)
	assert.Nil(t, got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepo_Get_Found(t *testing.T) {
	repo, mock, cleanup := newPermissionRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"account", "project", "can_download", "can_upload", "can_delete"}).
		AddRow(1, 2, true, false, true)
	mock.ExpectQuery(`SELECT account, project`).
		WithArgs(1, 2).
		WillReturnRows(rows)

	got, err := repo.Get(1, 2)
	assert.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 1, got.Account)
	assert.Equal(t, 2, got.Project)
	assert.True(t, got.CanDownload)
	assert.False(t, got.CanUpload)
	assert.True(t, got.CanDelete)
}

func TestPermissionRepo_Get_QueryError(t *testing.T) {
	repo, mock, cleanup := newPermissionRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT account, project`).
		WithArgs(1, 2).
		WillReturnError(fmt.Errorf("boom"))

	got, err := repo.Get(1, 2)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestPermissionRepo_Set_Upserts(t *testing.T) {
	repo, mock, cleanup := newPermissionRepo(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "permission"`)).
		WithArgs(1, 2, true, true, false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Set(types.Permission{Account: 1, Project: 2, CanDownload: true, CanUpload: true, CanDelete: false})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepo_Set_Error(t *testing.T) {
	repo, mock, cleanup := newPermissionRepo(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "permission"`)).
		WithArgs(1, 2, true, false, false).
		WillReturnError(fmt.Errorf("nope"))

	err := repo.Set(types.Permission{Account: 1, Project: 2, CanDownload: true})
	assert.Error(t, err)
}

func TestPermissionRepo_Delete(t *testing.T) {
	repo, mock, cleanup := newPermissionRepo(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "permission"`)).
		WithArgs(7, 9).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(7, 9)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepo_ListByProject(t *testing.T) {
	repo, mock, cleanup := newPermissionRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"account", "project", "can_download", "can_upload", "can_delete", "display_name"}).
		AddRow(5, 9, true, true, true, "Alice").
		AddRow(6, 9, true, false, false, "Bob")

	mock.ExpectQuery(`SELECT pe.account, pe.project`).
		WithArgs(9).
		WillReturnRows(rows)

	got, err := repo.ListByProject(9)
	assert.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "Alice", got[0].DisplayName)
	assert.Equal(t, 5, got[0].Account)
	assert.Equal(t, "Bob", got[1].DisplayName)
	assert.False(t, got[1].CanUpload)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPermissionRepo_ListByProject_Empty(t *testing.T) {
	repo, mock, cleanup := newPermissionRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"account", "project", "can_download", "can_upload", "can_delete", "display_name"})
	mock.ExpectQuery(`SELECT pe.account, pe.project`).
		WithArgs(9).
		WillReturnRows(rows)

	got, err := repo.ListByProject(9)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Len(t, got, 0)
}

func TestPermissionRepo_ListByProject_QueryError(t *testing.T) {
	repo, mock, cleanup := newPermissionRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT pe.account, pe.project`).
		WithArgs(9).
		WillReturnError(fmt.Errorf("boom"))

	got, err := repo.ListByProject(9)
	assert.Error(t, err)
	assert.Nil(t, got)
}
