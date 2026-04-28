package internal

import (
	"fmt"
	"database/sql"

	"packster/internal/logging"
	"packster/pkg/config"
	"packster/pkg/types"
)

var Hosts map[string]types.Host

func FetchHosts(cfg config.Config, conn *sql.DB) error {
	if Hosts == nil {
		Hosts = make(map[string]types.Host)
	}

	if conn == nil {
		return fmt.Errorf("Sql conn is nil")
	}

	if cfg.Gitlab != nil {
		for url, data := range cfg.Gitlab.Hosts {
			logging.Log.Debugf("Loading: %s", url)

			orgs, id, err := FetchOrgsByHostUrl(conn, url)
			if err != nil {
				logging.Log.Debugf("skiping %s", url)
				logging.Log.Error(err)
				continue
			}

			host := types.Host{
				Id: 		   id,
				Url:           url,
				Type:          types.Gitlab,
				ApplicationId: data.ApplicationId,
				Secret:        data.Secret,
				Orgs: 		   orgs,
			}

			Hosts[url] = host
		}
	}

	return nil
}

func HostByID(id int) (*types.Host, bool) {
	for _, v := range Hosts {
		if v.Id == id {
			h := v
			return &h, true
		}
	}
	return nil, false
}

func FetchOrgsByHostUrl(db *sql.DB, url string) ([]int, int, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM "host" WHERE url = $1)`, url).Scan(&exists)
	if err != nil {
		return nil, -1, err
	}

	if !exists {
		return nil, -1, fmt.Errorf("host not found: %s", url)
	}

	const query = `
		SELECT o.id, h.id
		FROM "org" o
		JOIN "host" h ON o.host = h.id
		WHERE h.url = $1
	`
	rows, err := db.Query(query, url)
	if err != nil {
		return nil, -1, err
	}
	defer rows.Close()

	var orgs []int
	var hostId int

	for rows.Next() {
		var id int
		if err := rows.Scan(&id, &hostId); err != nil {
			return nil, -1, err
		}
		orgs = append(orgs, id)
	}
	return orgs, hostId, rows.Err()
}
