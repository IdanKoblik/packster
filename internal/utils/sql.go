package utils

import (
	"database/sql"
	"packster/internal/logging"
)

func HostExists(url string, sqlConn *sql.DB) (bool, int) {
	if sqlConn == nil {
		return false, -1
	}

	var id int
	err := sqlConn.QueryRow(
		`SELECT id FROM host WHERE url=$1`,
		url,
	).Scan(&id)

	if err == sql.ErrNoRows {
		return false, -1
	} else if err != nil {
		logging.Log.Error(err)
		return false, -1
	}

	return true, id
}
