package redis

import (
	"testing"
	"path/filepath"

	"artifactor/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestOpenConnection_MissingConfig(t *testing.T) {
	err := OpenConnection(nil)
	assert.Nil(t, Client)
	assert.EqualError(t, err, "Missing redis config")
}

func TestOpenConnection_Sucess(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := config.ParseConfig(path)
	assert.NoError(t, err)

	err = OpenConnection(&cfg.Redis)
	assert.NoError(t, err)
	assert.NotNil(t, Client)

	Client.Close()
}

func TestOpenConnection_Invalid(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := config.ParseConfig(path)
	assert.NoError(t, err)

	cfg.Redis.Addr = "Invalid"
	err = OpenConnection(&cfg.Redis)
	assert.Error(t, err)
	assert.Nil(t, Client)
}
