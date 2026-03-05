package repository

import (
	"context"
	"path/filepath"
	"testing"

	"artifactor/internal/config"
	intsql "artifactor/internal/sql"
	intredis "artifactor/internal/redis"
	"artifactor/pkg/requests"
	"artifactor/pkg/users"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRepository(t *testing.T) (*AuthRepository, func()) {
	t.Helper()
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := config.ParseConfig(path)
	require.NoError(t, err)

	err = intsql.OpenConnection(&cfg.Sql)
	if err != nil {
		t.Skipf("skipping: could not connect to postgres: %v", err)
	}

	err = intredis.OpenConnection(&cfg.Redis)
	if err != nil {
		intsql.Conn.Close(context.Background())
		t.Skipf("skipping: could not connect to redis: %v", err)
	}

	repo := NewAuthRepository(intredis.Client, intsql.Conn)

	cleanup := func() {
		intsql.Conn.Close(context.Background())
		intredis.Client.Close()
	}

	return repo, cleanup
}

func deleteUser(t *testing.T, username string) {
	t.Helper()
	_, err := intsql.Conn.Exec(ctx, `DELETE FROM users WHERE username = $1`, username)
	require.NoError(t, err)
}

func newTestRequest(username string) *requests.RegisterRequest {
	return &requests.RegisterRequest{
		Username: username,
		Name:     "Test User",
		Mail:     "test@example.com",
		Password: "password123",
		Permissions: users.UserPermissions{
			Upload: true,
			Delete: false,
		},
	}
}

func TestCreateUser(t *testing.T) {
	repo, cleanup := setupRepository(t)
	defer cleanup()

	username := "repo_test_create"
	deleteUser(t, username)
	defer deleteUser(t, username)

	err := repo.CreateUser(newTestRequest(username))
	assert.NoError(t, err)
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	repo, cleanup := setupRepository(t)
	defer cleanup()

	username := "repo_test_duplicate"
	deleteUser(t, username)
	defer deleteUser(t, username)

	req := newTestRequest(username)
	require.NoError(t, repo.CreateUser(req))

	err := repo.CreateUser(req)
	assert.Error(t, err)
}

func TestUserExists_NotFound(t *testing.T) {
	repo, cleanup := setupRepository(t)
	defer cleanup()

	username := "repo_test_nonexistent_user_xyz"
	deleteUser(t, username)
	intredis.Client.Del(ctx, repo.getCacheKey(username))

	exists, err := repo.UserExists(username)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserExists_FoundInDB(t *testing.T) {
	repo, cleanup := setupRepository(t)
	defer cleanup()

	username := "repo_test_exists"
	deleteUser(t, username)
	defer deleteUser(t, username)
	intredis.Client.Del(ctx, repo.getCacheKey(username))

	require.NoError(t, repo.CreateUser(newTestRequest(username)))

	intredis.Client.Del(ctx, repo.getCacheKey(username))

	exists, err := repo.UserExists(username)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestUserExists_CacheHit(t *testing.T) {
	repo, cleanup := setupRepository(t)
	defer cleanup()

	username := "repo_test_cache"
	deleteUser(t, username)
	defer deleteUser(t, username)
	intredis.Client.Del(ctx, repo.getCacheKey(username))

	require.NoError(t, repo.CreateUser(newTestRequest(username)))

	exists, err := repo.UserExists(username)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.UserExists(username)
	assert.NoError(t, err)
	assert.True(t, exists)
}
