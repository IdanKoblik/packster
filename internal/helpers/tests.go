package helpers

import (
	"context"
	internalconfig "packster/internal/config"
	internalmysql "packster/internal/mysql"
	internalredis "packster/internal/redis"
	"packster/internal/repository"
	"packster/internal/utils"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func SetupRepo(t *testing.T) (*repository.AuthRepository, func()) {
	t.Helper()
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := internalconfig.ParseConfig(path)
	require.NoError(t, err)

	redisClient, err := internalredis.OpenConnection(&cfg.Redis)
	require.NoError(t, err)

	db, err := internalmysql.OpenConnection(&cfg.MySQL)
	require.NoError(t, err)

	repo := repository.NewAuthRepository(redisClient, db, &cfg)

	cleanup := func() {
		redisClient.Close()
		db.Close()
	}

	return repo, cleanup
}

// SetupProductRepo sets up a ProductRepository for integration tests.
// Any plaintext tokens passed are pre-registered in the database so that
// CreateProduct can look them up by hash.
func SetupProductRepo(t *testing.T, tokens ...string) (*repository.ProductRepository, func()) {
	t.Helper()
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := internalconfig.ParseConfig(path)
	require.NoError(t, err)

	db, err := internalmysql.OpenConnection(&cfg.MySQL)
	require.NoError(t, err)

	ctx := context.Background()
	var insertedPrincipalIDs []int64
	for _, tok := range tokens {
		hash := utils.Hash(tok)

		// Check if the token already exists (e.g. from a previous failed test run).
		var existing int64
		_ = db.QueryRowContext(ctx, "SELECT id FROM api_tokens WHERE token_hash = ?", hash).Scan(&existing)
		if existing != 0 {
			continue
		}

		result, err := db.ExecContext(ctx,
			"INSERT INTO principals (type, admin) VALUES ('token', FALSE)")
		require.NoError(t, err)
		principalID, err := result.LastInsertId()
		require.NoError(t, err)

		_, err = db.ExecContext(ctx,
			"INSERT INTO api_tokens (id, token_hash) VALUES (?, ?)", principalID, hash)
		require.NoError(t, err)

		insertedPrincipalIDs = append(insertedPrincipalIDs, principalID)
	}

	repo := repository.NewProductRepository(db, &cfg)

	cleanup := func() {
		for _, id := range insertedPrincipalIDs {
			db.ExecContext(ctx, "DELETE FROM principals WHERE id = ?", id)
		}
		db.Close()
	}

	return repo, cleanup
}
