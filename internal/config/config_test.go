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

		sql := cfg.Sql
		assert.Equal(t, "localhost", sql.Host)
		assert.Equal(t, uint16(5432), sql.Port)
		assert.Equal(t, "postgres", sql.DB)
		assert.Equal(t, "root", sql.User)
		assert.Equal(t, "root", sql.Password)
		assert.False(t, sql.SSL)
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
