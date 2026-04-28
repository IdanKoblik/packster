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

func newProductRepo(t *testing.T) (*ProductRepo, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return &ProductRepo{SqlDB: db}, mock, func() { db.Close() }
}

func TestProductRepo_Create(t *testing.T) {
	repo, mock, cleanup := newProductRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "product"`)).
		WithArgs("spigot", 5, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))

	p, err := repo.Create(5, "spigot")
	assert.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, 11, p.ID)
	assert.Equal(t, "spigot", p.Name)
	assert.Equal(t, 5, p.Project)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepo_Create_Error(t *testing.T) {
	repo, mock, cleanup := newProductRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "product"`)).
		WithArgs("spigot", 5, sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("dup"))

	p, err := repo.Create(5, "spigot")
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestProductRepo_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := newProductRepo(t)
	defer cleanup()

	mock.ExpectQuery(`WHERE id=`).
		WithArgs(11).
		WillReturnError(sql.ErrNoRows)

	p, err := repo.GetByID(11)
	assert.NoError(t, err)
	assert.Nil(t, p)
}

func TestProductRepo_GetByID_Found(t *testing.T) {
	repo, mock, cleanup := newProductRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "project", "created_at"}).
		AddRow(11, "spigot", 5, time.Time{})
	mock.ExpectQuery(`WHERE id=`).
		WithArgs(11).
		WillReturnRows(rows)

	p, err := repo.GetByID(11)
	assert.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, "spigot", p.Name)
}

func TestProductRepo_GetByName_NotFound(t *testing.T) {
	repo, mock, cleanup := newProductRepo(t)
	defer cleanup()

	mock.ExpectQuery(`WHERE project=\$1 AND name=\$2`).
		WithArgs(5, "spigot").
		WillReturnError(sql.ErrNoRows)

	p, err := repo.GetByName(5, "spigot")
	assert.NoError(t, err)
	assert.Nil(t, p)
}

func TestProductRepo_GetByName_Found(t *testing.T) {
	repo, mock, cleanup := newProductRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "project", "created_at"}).
		AddRow(11, "spigot", 5, time.Time{})
	mock.ExpectQuery(`WHERE project=\$1 AND name=\$2`).
		WithArgs(5, "spigot").
		WillReturnRows(rows)

	p, err := repo.GetByName(5, "spigot")
	assert.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, 11, p.ID)
}

func TestProductRepo_ListByProject(t *testing.T) {
	repo, mock, cleanup := newProductRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "project", "created_at"}).
		AddRow(1, "a", 5, time.Time{}).
		AddRow(2, "b", 5, time.Time{})
	mock.ExpectQuery(`FROM "product" WHERE project=`).
		WithArgs(5).
		WillReturnRows(rows)

	got, err := repo.ListByProject(5)
	assert.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestProductRepo_ListByProject_QueryError(t *testing.T) {
	repo, mock, cleanup := newProductRepo(t)
	defer cleanup()

	mock.ExpectQuery(`FROM "product" WHERE project=`).
		WithArgs(5).
		WillReturnError(fmt.Errorf("boom"))

	got, err := repo.ListByProject(5)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestProductRepo_Delete(t *testing.T) {
	repo, mock, cleanup := newProductRepo(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "product"`)).
		WithArgs(11).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(11)
	assert.NoError(t, err)
}
