package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"artifactor/pkg/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type testRedisClient struct {
	getResult string
	getErr    error
	setErr    error
	delResult int64
	delErr    error
}

func (c *testRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return redis.NewStringResult(c.getResult, c.getErr)
}

func (c *testRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return redis.NewStatusResult("", c.setErr)
}

func (c *testRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return redis.NewIntResult(c.delResult, c.delErr)
}

type testSQLClient struct {
	execErr   error
	scanValue interface{}
	scanErr   error
}

func (c *testSQLClient) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, c.execErr
}

func (c *testSQLClient) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return &testRow{value: c.scanValue, err: c.scanErr}
}

type testRowWithExec struct {
	err error
}

func (r *testRowWithExec) Scan(dest ...interface{}) error {
	return r.err
}

type testSQLClientWithExec struct {
	execErr error
	row     *testRowWithExec
}

func (c *testSQLClientWithExec) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, c.execErr
}

func (c *testSQLClientWithExec) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return c.row
}

type testRow struct {
	value interface{}
	err   error
}

func (r *testRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	return nil
}

func TestGetCacheKey(t *testing.T) {
	repo := &AuthRepository{}

	tests := []struct {
		token    string
		expected string
	}{
		{"test-token", "token:test-token"},
		{"abc123", "token:abc123"},
		{"", "token:"},
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			key := repo.getCacheKey(tt.token)
			assert.Equal(t, tt.expected, key)
		})
	}
}

func TestNewAuthRepository(t *testing.T) {
	repo := NewAuthRepository(nil, nil)
	assert.NotNil(t, repo)
	assert.Nil(t, repo.Rdb)
	assert.Nil(t, repo.SqlClient)
	assert.Equal(t, 5*time.Minute, repo.CacheTTL)
}

func TestTokenExists_CacheHit(t *testing.T) {
	repo := &AuthRepository{
		Rdb: &testRedisClient{getResult: "true", getErr: nil},
	}

	exists, err := repo.TokenExists("test-token")

	assert.True(t, exists)
	assert.NoError(t, err)
}

func TestTokenExists_CacheMiss_DBHit(t *testing.T) {
	repo := &AuthRepository{
		Rdb:       &testRedisClient{getResult: "", getErr: redis.Nil},
		SqlClient: &testSQLClient{scanValue: 1},
	}

	exists, err := repo.TokenExists("test-token")

	assert.True(t, exists)
	assert.NoError(t, err)
}

func TestTokenExists_NotFound(t *testing.T) {
	repo := &AuthRepository{
		Rdb:       &testRedisClient{getResult: "", getErr: redis.Nil},
		SqlClient: &testSQLClient{scanErr: pgx.ErrNoRows},
	}

	exists, err := repo.TokenExists("nonexistent")

	assert.False(t, exists)
	assert.NoError(t, err)
}

func TestTokenExists_DBError(t *testing.T) {
	repo := &AuthRepository{
		Rdb:       &testRedisClient{getResult: "", getErr: redis.Nil},
		SqlClient: &testSQLClient{scanErr: errors.New("db error")},
	}

	exists, err := repo.TokenExists("test-token")

	assert.False(t, exists)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestPruneToken_Success(t *testing.T) {
	repo := &AuthRepository{
		Rdb:       &testRedisClient{getResult: "true", delResult: 1},
		SqlClient: &testSQLClient{},
	}

	err := repo.PruneToken("test-token")

	assert.NoError(t, err)
}

func TestPruneToken_NotFound(t *testing.T) {
	repo := &AuthRepository{
		Rdb:       &testRedisClient{getResult: "", getErr: redis.Nil},
		SqlClient: &testSQLClient{scanErr: pgx.ErrNoRows},
	}

	err := repo.PruneToken("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exists")
}

func TestIsAdmin_Success(t *testing.T) {
	repo := &AuthRepository{
		Rdb:       &testRedisClient{getResult: "true"},
		SqlClient: &testSQLClient{},
	}

	_, err := repo.IsAdmin("test-token")

	assert.NoError(t, err)
}

func TestIsAdmin_NotFound(t *testing.T) {
	repo := &AuthRepository{
		Rdb:       &testRedisClient{getResult: "", getErr: redis.Nil},
		SqlClient: &testSQLClient{scanErr: pgx.ErrNoRows},
	}

	_, err := repo.IsAdmin("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFetchToken_Success(t *testing.T) {
	repo := &AuthRepository{
		Rdb:       &testRedisClient{getResult: "true"},
		SqlClient: &testSQLClient{},
	}

	token, err := repo.FetchToken("test-token")

	assert.NotNil(t, token)
	assert.NoError(t, err)
}

func TestFetchToken_NotFound(t *testing.T) {
	repo := &AuthRepository{
		Rdb:       &testRedisClient{getResult: "", getErr: redis.Nil},
		SqlClient: &testSQLClient{scanErr: pgx.ErrNoRows},
	}

	token, err := repo.FetchToken("nonexistent")

	assert.Nil(t, token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exists")
}

func TestCreateToken_Success(t *testing.T) {
	repo := &AuthRepository{
		Rdb:       &testRedisClient{},
		SqlClient: &testSQLClientWithExec{},
		CacheTTL:  time.Minute,
	}

	token, err := repo.CreateToken(&http.CreateRequest{Admin: true, Upload: true, Delete: false})

	assert.NotEmpty(t, token)
	assert.NoError(t, err)
}

func TestCreateToken_SQLError(t *testing.T) {
	repo := &AuthRepository{
		Rdb:       &testRedisClient{},
		SqlClient: &testSQLClientWithExec{execErr: errors.New("sql error")},
		CacheTTL:  time.Minute,
	}

	token, err := repo.CreateToken(&http.CreateRequest{Admin: true})

	assert.Empty(t, token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sql error")
}
