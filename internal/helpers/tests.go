package helpers

import (
	internalconfig "packster/internal/config"
	internalmongo "packster/internal/mongo"
	internalredis "packster/internal/redis"
	"packster/internal/repository"
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

	mongoClient, err := internalmongo.OpenConnection(&cfg.Mongo)
	require.NoError(t, err)

	repo := repository.NewAuthRepository(redisClient, mongoClient, &cfg)

	cleanup := func() {
		redisClient.Close()
		mongoClient.Disconnect(nil)
	}

	return repo, cleanup
}

func SetupProductRepo(t *testing.T) (*repository.ProductRepository, func()) {
	t.Helper()
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := internalconfig.ParseConfig(path)
	require.NoError(t, err)

	mongoClient, err := internalmongo.OpenConnection(&cfg.Mongo)
	require.NoError(t, err)

	repo := repository.NewProductRepository(mongoClient, &cfg)

	cleanup := func() {
		mongoClient.Disconnect(nil)
	}

	return repo, cleanup
}
