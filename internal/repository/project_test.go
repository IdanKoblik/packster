package repository

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newProjectRepo(t *testing.T) (*ProjectRepo, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return &ProjectRepo{SqlDB: db}, mock, func() { db.Close() }
}

func TestProjectRepo_Import_Success(t *testing.T) {
	repo, mock, cleanup := newProjectRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "project"`)).
		WithArgs(1, 99, 7, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "permission"`)).
		WithArgs(7, 42).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	p, err := repo.Import(7, 1, 99)
	assert.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, 42, p.ID)
	assert.Equal(t, 1, p.Host)
	assert.Equal(t, 99, p.Repository)
	assert.Equal(t, 7, p.Owner)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepo_Import_ProjectInsertFails_Rollback(t *testing.T) {
	repo, mock, cleanup := newProjectRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "project"`)).
		WithArgs(1, 99, 7, sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("boom"))
	mock.ExpectRollback()

	p, err := repo.Import(7, 1, 99)
	assert.Error(t, err)
	assert.Nil(t, p)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepo_Import_PermInsertFails_Rollback(t *testing.T) {
	repo, mock, cleanup := newProjectRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "project"`)).
		WithArgs(1, 99, 7, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "permission"`)).
		WithArgs(7, 42).
		WillReturnError(fmt.Errorf("perm fail"))
	mock.ExpectRollback()

	p, err := repo.Import(7, 1, 99)
	assert.Error(t, err)
	assert.Nil(t, p)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepo_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := newProjectRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, host, repository, owner, created_at FROM "project" WHERE id=`).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	p, err := repo.GetByID(99)
	assert.NoError(t, err)
	assert.Nil(t, p)
}

func TestProjectRepo_GetByID_Found(t *testing.T) {
	repo, mock, cleanup := newProjectRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "host", "repository", "owner", "created_at"}).
		AddRow(42, 1, 99, 7, time.Time{})

	mock.ExpectQuery(`SELECT id, host, repository, owner, created_at FROM "project" WHERE id=`).
		WithArgs(42).
		WillReturnRows(rows)

	p, err := repo.GetByID(42)
	assert.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, 42, p.ID)
	assert.Equal(t, 7, p.Owner)
}

func TestProjectRepo_GetByHostRepository_NotFound(t *testing.T) {
	repo, mock, cleanup := newProjectRepo(t)
	defer cleanup()

	mock.ExpectQuery(`WHERE host=\$1 AND repository=\$2`).
		WithArgs(1, 99).
		WillReturnError(sql.ErrNoRows)

	p, err := repo.GetByHostRepository(1, 99)
	assert.NoError(t, err)
	assert.Nil(t, p)
}

func TestProjectRepo_GetByHostRepository_Found(t *testing.T) {
	repo, mock, cleanup := newProjectRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "host", "repository", "owner", "created_at"}).
		AddRow(42, 1, 99, 7, time.Time{})

	mock.ExpectQuery(`WHERE host=\$1 AND repository=\$2`).
		WithArgs(1, 99).
		WillReturnRows(rows)

	p, err := repo.GetByHostRepository(1, 99)
	assert.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, 42, p.ID)
}

func TestProjectRepo_ListAccessible(t *testing.T) {
	repo, mock, cleanup := newProjectRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "host", "repository", "owner", "created_at"}).
		AddRow(1, 1, 11, 5, time.Time{}).
		AddRow(2, 1, 22, 5, time.Time{})

	mock.ExpectQuery(`JOIN "permission" pe ON pe.project = p.id`).
		WithArgs(5).
		WillReturnRows(rows)

	got, err := repo.ListAccessible(5)
	assert.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, 1, got[0].ID)
	assert.Equal(t, 2, got[1].ID)
}

func TestProjectRepo_ListAccessible_QueryError(t *testing.T) {
	repo, mock, cleanup := newProjectRepo(t)
	defer cleanup()

	mock.ExpectQuery(`JOIN "permission"`).
		WithArgs(5).
		WillReturnError(fmt.Errorf("boom"))

	got, err := repo.ListAccessible(5)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestProjectRepo_Delete_Success(t *testing.T) {
	repo, mock, cleanup := newProjectRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	pathRows := sqlmock.NewRows([]string{"path"}).
		AddRow("/var/blobs/a").
		AddRow("/var/blobs/b")
	mock.ExpectQuery(`SELECT v.path`).
		WithArgs(42).
		WillReturnRows(pathRows)

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "version"`)).
		WithArgs(42).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "token_access"`)).
		WithArgs(42).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "product"`)).
		WithArgs(42).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "permission"`)).
		WithArgs(42).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "project"`)).
		WithArgs(42).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	paths, err := repo.Delete(42)
	assert.NoError(t, err)
	assert.Equal(t, []string{"/var/blobs/a", "/var/blobs/b"}, paths)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepo_Delete_NoVersions(t *testing.T) {
	repo, mock, cleanup := newProjectRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT v.path`).
		WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"path"}))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "version"`)).WithArgs(42).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "token_access"`)).WithArgs(42).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "product"`)).WithArgs(42).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "permission"`)).WithArgs(42).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "project"`)).WithArgs(42).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	paths, err := repo.Delete(42)
	assert.NoError(t, err)
	assert.Empty(t, paths)
}

func TestProjectRepo_Delete_PathQueryFails_Rollback(t *testing.T) {
	repo, mock, cleanup := newProjectRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT v.path`).
		WithArgs(42).
		WillReturnError(fmt.Errorf("boom"))
	mock.ExpectRollback()

	paths, err := repo.Delete(42)
	assert.Error(t, err)
	assert.Nil(t, paths)
}

func TestProjectRepo_Delete_StepFails_Rollback(t *testing.T) {
	repo, mock, cleanup := newProjectRepo(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT v.path`).
		WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"path"}))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "version"`)).
		WithArgs(42).
		WillReturnError(fmt.Errorf("nope"))
	mock.ExpectRollback()

	paths, err := repo.Delete(42)
	assert.Error(t, err)
	assert.Nil(t, paths)
}
