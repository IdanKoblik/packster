package repository

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newVersionRepo(t *testing.T) (*VersionRepo, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return &VersionRepo{SqlDB: db}, mock, func() { db.Close() }
}

func TestVersionRepo_Create(t *testing.T) {
	repo, mock, cleanup := newVersionRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "version"`)).
		WithArgs("1.0.0", "/blob/path", "abc", 11).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))

	v, err := repo.Create(11, "1.0.0", "/blob/path", "abc")
	assert.NoError(t, err)
	require.NotNil(t, v)
	assert.Equal(t, 99, v.ID)
	assert.Equal(t, "1.0.0", v.Name)
	assert.Equal(t, "abc", v.Checksum)
}

func TestVersionRepo_Create_Error(t *testing.T) {
	repo, mock, cleanup := newVersionRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "version"`)).
		WithArgs("1.0.0", "/p", "abc", 11).
		WillReturnError(fmt.Errorf("dup"))

	v, err := repo.Create(11, "1.0.0", "/p", "abc")
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestVersionRepo_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := newVersionRepo(t)
	defer cleanup()

	mock.ExpectQuery(`FROM "version" WHERE id=`).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	v, err := repo.GetByID(99)
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestVersionRepo_GetByID_Found(t *testing.T) {
	repo, mock, cleanup := newVersionRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "path", "checksum", "product"}).
		AddRow(99, "1.0.0", "/p", "abc", 11)
	mock.ExpectQuery(`FROM "version" WHERE id=`).
		WithArgs(99).
		WillReturnRows(rows)

	v, err := repo.GetByID(99)
	assert.NoError(t, err)
	require.NotNil(t, v)
	assert.Equal(t, 11, v.Product)
}

func TestVersionRepo_GetByName(t *testing.T) {
	repo, mock, cleanup := newVersionRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "path", "checksum", "product"}).
		AddRow(99, "1.0.0", "/p", "abc", 11)
	mock.ExpectQuery(`WHERE product=\$1 AND name=\$2`).
		WithArgs(11, "1.0.0").
		WillReturnRows(rows)

	v, err := repo.GetByName(11, "1.0.0")
	assert.NoError(t, err)
	require.NotNil(t, v)
	assert.Equal(t, 99, v.ID)
}

func TestVersionRepo_GetByName_NotFound(t *testing.T) {
	repo, mock, cleanup := newVersionRepo(t)
	defer cleanup()

	mock.ExpectQuery(`WHERE product=\$1 AND name=\$2`).
		WithArgs(11, "1.0.0").
		WillReturnError(sql.ErrNoRows)

	v, err := repo.GetByName(11, "1.0.0")
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestVersionRepo_ListByProduct(t *testing.T) {
	repo, mock, cleanup := newVersionRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "path", "checksum", "product"}).
		AddRow(1, "a", "/a", "1", 11).
		AddRow(2, "b", "/b", "2", 11)
	mock.ExpectQuery(`FROM "version" WHERE product=`).
		WithArgs(11).
		WillReturnRows(rows)

	got, err := repo.ListByProduct(11)
	assert.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestVersionRepo_ListByProduct_Error(t *testing.T) {
	repo, mock, cleanup := newVersionRepo(t)
	defer cleanup()

	mock.ExpectQuery(`FROM "version" WHERE product=`).
		WithArgs(11).
		WillReturnError(fmt.Errorf("boom"))

	got, err := repo.ListByProduct(11)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestVersionRepo_Delete(t *testing.T) {
	repo, mock, cleanup := newVersionRepo(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "version"`)).
		WithArgs(99).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(99)
	assert.NoError(t, err)
}
