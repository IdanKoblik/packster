package sql

import (
	"testing"
	"context"
	"path/filepath"

	"artifactor/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestOpenConnection_Sucess(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "config.yml")
	cfg, err := config.ParseConfig(path)
	assert.NoError(t, err)

	conn, err := OpenConnection(&cfg.Sql)
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	conn.Close(context.Background())
}

func TestOpenConnection_Invalid(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := config.ParseConfig(path)
	assert.NoError(t, err)

	conn, err := OpenConnection(&cfg.Sql)
	assert.Error(t, err)
	assert.Nil(t, conn)
}
