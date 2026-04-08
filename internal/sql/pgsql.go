package sql

import (
	"time"
    "database/sql"
	"packster/pkg/config"

    "github.com/lib/pq"
)

var PgsqlConn *sql.DB

func OpenPgsqlConnection(cfg *config.PgsqlConfig) error {
	sslMode := pq.SSLModeDisable
	if cfg.SSL {
		sslMode = pq.SSLModeRequire
	}

	pgsql := pq.Config{
		Host: cfg.Host,
		Port: cfg.Port,
		Database: cfg.DB,
		User: cfg.User,
		Password: cfg.Password,
		SSLMode: sslMode,
		ConnectTimeout: 5 * time.Second,
	}

	c, err := pq.NewConnectorConfig(pgsql)
	if err != nil {
		return err
	}

	PgsqlConn = sql.OpenDB(c)

	err = PgsqlConn.Ping()
	if err != nil {
		return err
	}

	return nil
}
