package sql

import (
	"time"
    "database/sql"

	"packster/pkg/config"
	"packster/internal/logging"
	"packster/internal/utils"

	"github.com/lib/pq"
)

var PgsqlConn *sql.DB

func OpenPgsqlConnection(cfg *config.PgsqlConfig) (*sql.DB, error) {
	logging.Log.Info("Connecting to pgsql db:")
	logging.Log.Infof("Host: %s", cfg.Host)
	logging.Log.Infof("Port: %d", cfg.Port)
	logging.Log.Infof("Username: %s", cfg.User)
	logging.Log.Infof("Password: %s", utils.GenerateMask())

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
		return nil, err
	}

	PgsqlConn = sql.OpenDB(c)

	err = PgsqlConn.Ping()
	if err != nil {
		return nil, err
	}

	return PgsqlConn, nil
}
