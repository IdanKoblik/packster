package helpers

import (
	internalconfig "artifactor/internal/config"
	internalmongo "artifactor/internal/mongo"
	internalredis "artifactor/internal/redis"
	"artifactor/internal/repository"
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
