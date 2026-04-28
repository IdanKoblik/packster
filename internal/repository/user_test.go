package repository

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"

	"packster/internal"
	"packster/pkg/types"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRepo(t *testing.T) (*UserRepo, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return &UserRepo{SqlDB: db}, mock, func() { db.Close() }
}

func withHosts(hosts map[string]types.Host) func() {
	prev := internal.Hosts
	internal.Hosts = hosts
	return func() { internal.Hosts = prev }
}

func TestIntersectOrgs(t *testing.T) {
	tests := []struct {
		name     string
		user     []int
		host     []int
		expected []int
	}{
		{"both empty", nil, nil, []int{}},
		{"user empty", nil, []int{1, 2}, []int{}},
		{"host empty", []int{1, 2}, nil, []int{}},
		{"no overlap", []int{1, 2}, []int{3, 4}, []int{}},
		{"full overlap", []int{1, 2}, []int{1, 2}, []int{1, 2}},
		{"partial overlap", []int{1, 2, 3}, []int{2, 3, 4}, []int{2, 3}},
		{"preserves user order", []int{3, 1, 2}, []int{1, 2, 3}, []int{3, 1, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intersectOrgs(tt.user, tt.host)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestUserExists_NotFound(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT a.display_name`).
		WithArgs("alice", 42, 1).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.UserExists("alice", 42, 1)
	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserExists_Found(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"display_name", "account", "username", "sso_id", "host"}).
		AddRow("Alice", 7, "alice", 42, "1")

	mock.ExpectQuery(`SELECT a.display_name`).
		WithArgs("alice", 42, 1).
		WillReturnRows(rows)

	user, err := repo.UserExists("alice", 42, 1)
	assert.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, 7, user.ID)
	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, "Alice", user.DisplayName)
	assert.Equal(t, 42, user.SsoID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserExists_QueryError(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT a.display_name`).
		WithArgs("alice", 42, 1).
		WillReturnError(fmt.Errorf("boom"))

	user, err := repo.UserExists("alice", 42, 1)
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestCreateUser_HostNotFound(t *testing.T) {
	repo, _, cleanup := newRepo(t)
	defer cleanup()
	defer withHosts(map[string]types.Host{})()

	user, err := repo.CreateUser("alice", "https://missing", 1, []int{1})
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "No such host")
}

func TestCreateUser_NoMatchingOrgs(t *testing.T) {
	repo, _, cleanup := newRepo(t)
	defer cleanup()

	hostURL := "https://gitlab.example.com"
	defer withHosts(map[string]types.Host{
		hostURL: {Id: 1, Url: hostURL, Orgs: []int{10, 20}},
	})()

	user, err := repo.CreateUser("alice", hostURL, 42, []int{99})
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "no orgs matching host")
}

func TestCreateUser_ExistingUser(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	hostURL := "https://gitlab.example.com"
	defer withHosts(map[string]types.Host{
		hostURL: {Id: 1, Url: hostURL, Orgs: []int{10, 20}},
	})()

	rows := sqlmock.NewRows([]string{"display_name", "account", "username", "sso_id", "host"}).
		AddRow("Alice", 5, "alice", 42, "1")
	mock.ExpectQuery(`SELECT a.display_name`).
		WithArgs("alice", 42, 1).
		WillReturnRows(rows)

	user, err := repo.CreateUser("alice", hostURL, 42, []int{10, 99})
	assert.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, 5, user.ID)
	assert.Equal(t, []int{10}, user.Orgs)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_NewUser(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	hostURL := "https://gitlab.example.com"
	defer withHosts(map[string]types.Host{
		hostURL: {Id: 1, Url: hostURL, Orgs: []int{10, 20}},
	})()

	mock.ExpectQuery(`SELECT a.display_name`).
		WithArgs("alice", 42, 1).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "user"`)).
		WithArgs("alice", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))
	mock.ExpectExec(`INSERT INTO auth`).
		WithArgs(99, "alice", 42, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user, err := repo.CreateUser("alice", hostURL, 42, []int{10, 20, 99})
	assert.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, 99, user.ID)
	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, "alice", user.DisplayName)
	assert.Equal(t, 42, user.SsoID)
	assert.Equal(t, hostURL, user.Host)
	assert.Equal(t, []int{10, 20}, user.Orgs)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_InsertUserFails_Rollback(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	hostURL := "https://gitlab.example.com"
	defer withHosts(map[string]types.Host{
		hostURL: {Id: 1, Url: hostURL, Orgs: []int{10}},
	})()

	mock.ExpectQuery(`SELECT a.display_name`).
		WithArgs("alice", 42, 1).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "user"`)).
		WithArgs("alice", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert user failed"))
	mock.ExpectRollback()

	user, err := repo.CreateUser("alice", hostURL, 42, []int{10})
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_InsertAuthFails_Rollback(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	hostURL := "https://gitlab.example.com"
	defer withHosts(map[string]types.Host{
		hostURL: {Id: 1, Url: hostURL, Orgs: []int{10}},
	})()

	mock.ExpectQuery(`SELECT a.display_name`).
		WithArgs("alice", 42, 1).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "user"`)).
		WithArgs("alice", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectExec(`INSERT INTO auth`).
		WithArgs(7, "alice", 42, 1).
		WillReturnError(fmt.Errorf("insert auth failed"))
	mock.ExpectRollback()

	user, err := repo.CreateUser("alice", hostURL, 42, []int{10})
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserExistsByID_True(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(7).
		WillReturnRows(rows)

	got, err := repo.UserExistsByID(7)
	assert.NoError(t, err)
	assert.True(t, got)
}

func TestUserExistsByID_False(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(7).
		WillReturnRows(rows)

	got, err := repo.UserExistsByID(7)
	assert.NoError(t, err)
	assert.False(t, got)
}

func TestUserExistsByID_Error(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(7).
		WillReturnError(fmt.Errorf("boom"))

	got, err := repo.UserExistsByID(7)
	assert.Error(t, err)
	assert.False(t, got)
}

func TestSearchByName_Found(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "display_name", "username", "sso_id"}).
		AddRow(2, "Bob", "bob", 200).
		AddRow(3, "Bobby", "bobby", 201)

	mock.ExpectQuery(`FROM "user" u`).
		WithArgs(1, "Bo%", 7).
		WillReturnRows(rows)

	users, err := repo.SearchByName(1, "Bo", 7)
	assert.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, "Bob", users[0].DisplayName)
	assert.Equal(t, "Bobby", users[1].DisplayName)
}

func TestSearchByName_Empty(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "display_name", "username", "sso_id"})
	mock.ExpectQuery(`FROM "user" u`).
		WithArgs(1, "X%", 7).
		WillReturnRows(rows)

	users, err := repo.SearchByName(1, "X", 7)
	assert.NoError(t, err)
	assert.Len(t, users, 0)
}

func TestSearchByName_QueryError(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	mock.ExpectQuery(`FROM "user" u`).
		WithArgs(1, "Bo%", 7).
		WillReturnError(fmt.Errorf("boom"))

	users, err := repo.SearchByName(1, "Bo", 7)
	assert.Error(t, err)
	assert.Nil(t, users)
}

func TestCreateUser_UserExistsQueryError(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	hostURL := "https://gitlab.example.com"
	defer withHosts(map[string]types.Host{
		hostURL: {Id: 1, Url: hostURL, Orgs: []int{10}},
	})()

	mock.ExpectQuery(`SELECT a.display_name`).
		WithArgs("alice", 42, 1).
		WillReturnError(fmt.Errorf("db down"))

	user, err := repo.CreateUser("alice", hostURL, 42, []int{10})
	assert.Error(t, err)
	assert.Nil(t, user)
}
