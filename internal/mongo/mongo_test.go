package mongo

import (
	"path/filepath"
	"testing"
	"time"

	internalconfig "packster/internal/config"
	pkgconfig "packster/pkg/config"

	"github.com/stretchr/testify/assert"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestOpenConnection_Success(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := internalconfig.ParseConfig(path)
	assert.NoError(t, err)

	client, err := OpenConnection(&cfg.Mongo)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	client.Disconnect(nil)
}

func TestOpenConnection_InvalidURI(t *testing.T) {
	cfg := &pkgconfig.MongoConfig{
		ConnectionString: "not-a-valid-uri",
	}

	client, err := OpenConnection(cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestCheckHealth_Success(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "example.yml")
	cfg, err := internalconfig.ParseConfig(path)
	assert.NoError(t, err)

	client, err := OpenConnection(&cfg.Mongo)
	assert.NoError(t, err)
	defer client.Disconnect(nil)

	err = CheckHealth(client)
	assert.NoError(t, err)
}

func TestCheckHealth_Unreachable(t *testing.T) {
	client, err := driver.Connect(
		options.Client().
			ApplyURI("mongodb://localhost:1/").
			SetServerSelectionTimeout(500 * time.Millisecond),
	)
	assert.NoError(t, err)
	defer client.Disconnect(nil)

	err = CheckHealth(client)
	assert.Error(t, err)
}
