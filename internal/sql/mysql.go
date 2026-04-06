package sql

import (
	"database/sql"
	"packster/pkg/config"

	_ "github.com/go-sql-driver/mysql"
)

var MysqlConn *sql.DB

func ConnectToMysql(cfg *config.MysqlConfig) error {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	MysqlConn = db
	return nil
}
