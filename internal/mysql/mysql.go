package mysql

import (
	"database/sql"
	"fmt"
	"packster/pkg/config"

	_ "github.com/go-sql-driver/mysql"
)

func OpenConnection(cfg *config.MySQLConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, err
	}

	if err := CheckHealth(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func CheckHealth(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("Missing mysql client")
	}

	return db.Ping()
}
