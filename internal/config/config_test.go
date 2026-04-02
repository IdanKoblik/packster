package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig_Success(t *testing.T) {
	files := [2]string{"example.yaml", "example.yml"}

	for _, file := range files {
		path := filepath.Join("..", "..", "fixtures", file)

		cfg, err := ParseConfig(path)
		assert.NoError(t, err)

		assert.Equal(t, 20, cfg.FileUploadLimit)
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
