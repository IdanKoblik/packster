package redis

import (
	"path/filepath"
	"testing"

	"artifactor/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestOpenConnection_MissingConfig(t *testing.T) {
	client, err := OpenConnection(nil)
	assert.Nil(t, client)
	assert.EqualError(t, err, "Missing redis config")
}

func TestOpenConnection_Sucess(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := config.ParseConfig(path)
	assert.NoError(t, err)

	client, err := OpenConnection(&cfg.Redis)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	client.Close()
}

func TestOpenConnection_Invalid(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := config.ParseConfig(path)
	assert.NoError(t, err)

	cfg.Redis.Addr = "Invalid"
	client, err := OpenConnection(&cfg.Redis)
	assert.Error(t, err)
	assert.Nil(t, client)
}
