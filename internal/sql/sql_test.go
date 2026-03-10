package sql

import (
	"context"
	"path/filepath"
	"testing"

	"artifactor/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestOpenConnection_MissingConfig(t *testing.T) {
	err := OpenConnection(nil)
	assert.Nil(t, Conn)
	assert.EqualError(t, err, "Missing pgsql config")
}

func TestCheckHealth_NotInitialized(t *testing.T) {
	origConn := Conn
	Conn = nil
	defer func() { Conn = origConn }()

	err := CheckHealth()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestOpenConnection_Sucess(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := config.ParseConfig(path)
	assert.NoError(t, err)

	err = OpenConnection(&cfg.Sql)
	assert.NoError(t, err)
	assert.NotNil(t, Conn)

	err = CheckHealth()
	assert.NoError(t, err)

	Conn.Close(context.Background())
}

func TestOpenConnection_Invalid(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := config.ParseConfig(path)
	assert.NoError(t, err)

	cfg.Sql.Addr = "invalid"
	err = OpenConnection(&cfg.Sql)
	assert.Error(t, err)
	assert.Nil(t, Conn)

	err = CheckHealth()
	assert.Error(t, err)
}
