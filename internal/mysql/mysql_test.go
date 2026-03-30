package mysql

import (
	"path/filepath"
	"testing"

	internalconfig "packster/internal/config"
	pkgconfig "packster/pkg/config"

	"github.com/stretchr/testify/assert"
)

func TestOpenConnection_Success(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := internalconfig.ParseConfig(path)
	assert.NoError(t, err)

	db, err := OpenConnection(&cfg.MySQL)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	db.Close()
}

func TestOpenConnection_InvalidDSN(t *testing.T) {
	cfg := &pkgconfig.MySQLConfig{
		DSN: "not-a-valid-dsn",
	}

	db, err := OpenConnection(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestCheckHealth_Success(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := internalconfig.ParseConfig(path)
	assert.NoError(t, err)

	db, err := OpenConnection(&cfg.MySQL)
	assert.NoError(t, err)
	defer db.Close()

	err = CheckHealth(db)
	assert.NoError(t, err)
}

func TestCheckHealth_Nil(t *testing.T) {
	err := CheckHealth(nil)
	assert.Error(t, err)
}
