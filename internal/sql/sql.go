package sql

import (
	"context"
	"net/url"
	"artifactor/pkg/config"

	"github.com/jackc/pgx/v5"
)

func OpenConnection(cfg *config.PgsqlConfig) (*pgx.Conn, error) {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.Username, cfg.Password),
		Host:   cfg.Addr,
		Path:   cfg.Database,
	}

	conn, err := pgx.Connect(context.Background(), u.String())
	if err != nil {
		return nil, err
	}

	return conn, nil
}
