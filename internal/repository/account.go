package repository

import (
	"time"
	"fmt"
	"database/sql"

	"packster/pkg/types"
	"packster/internal/utils"
)

type IAccountRepo interface {
	CreateAccount(request types.AuthRequest) error
}

type AccountRepo struct {
	SqlConn *sql.DB
}

func NewAccountRepo(sqlConn *sql.DB) *AccountRepo {
	return &AccountRepo{
		SqlConn: sqlConn,
	}
}

func (r *AccountRepo) CreateAccount(request types.AuthRequest) error {
	hostExists, hostId := utils.HostExists(request.Host, r.SqlConn)
	if !hostExists {
		return fmt.Errorf("%s isnt a valid host", request.Host)
	}

	exists, err := accountExists(request.Username, request.SsoId, hostId, r.SqlConn)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("This account already exists")
	}

	tx, err := r.SqlConn.Begin()
	if err != nil {
		return err
	}

	var accountID int
	err = tx.QueryRow(`INSERT INTO account (display_name, last_login, created_at) VALUES ($1, $2, $3) RETURNING id`,
		request.Username,
		time.Now(),
		time.Now(),
	).Scan(&accountID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`INSERT INTO auth (account, username, sso_id, host) VALUES ($1, $2, $3, $4)`,
		accountID,
		request.Username,
		request.SsoId,
		hostId,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func accountExists(username, sso string, host int, sqlConn *sql.DB) (bool, error) {
	var exists bool
	err := sqlConn.QueryRow(`SELECT EXISTS(SELECT 1 FROM auth WHERE username=$1 AND sso_id=$2 AND host=$3)`,
		username,
		sso,
		host,
	).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}
