package config

import (
	"testing"
	"path/filepath"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig_Success(t *testing.T) {
	files := [2]string{"example.yaml", "example.yml"}

	for _, file := range files {
		path := filepath.Join("..", "..", "fixtures", file)

		cfg, err := ParseConfig(path)
		assert.NoError(t, err)

		assert.Equal(t, 20, cfg.FileUploadLimit)
		assert.Equal(t, "key", cfg.SigningKey)

		assert.Equal(t, "username", cfg.Sql.Username)
		assert.Equal(t, "password", cfg.Sql.Password)
		assert.Equal(t, "localhost:5173", cfg.Sql.Addr)
		assert.Equal(t, "db", cfg.Sql.Database)

		assert.Equal(t, "localhost:6379", cfg.Redis.Addr)
		assert.Equal(t, "", cfg.Redis.Password)
		assert.Equal(t, 0, cfg.Redis.DB)
	}
}

func TestParseConfig_InvalidFile(t *testing.T) {
	path := "test.txt"

	_, err := ParseConfig(path)
	assert.EqualError(t, err, "Unsupported config file type")

	path = "invalid.yml"
	_, err = ParseConfig(path)
	assert.ErrorContains(t, err, "no such file")

	path = filepath.Join("..", "..", "fixtures", "invalid.yml")
	_, err = ParseConfig(path)
	assert.Error(t, err)
}
