package repository

import (
	"errors"
	"time"
	"fmt"
	"database/sql"

	"packster/internal"
	"packster/pkg/types"
)

type IUserRepo interface {
	CreateUser(username, host string, ssoId int, orgs []int) (*types.User, error)
	UserExists(username string, ssoId, host int) (*types.User, error)
	UserExistsByID(id int) (bool, error)
	PurgeUserData(userID int) ([]string, error)
	SearchByName(hostID int, query string, excludeID int) ([]types.User, error)
}

type UserRepo struct {
	SqlDB *sql.DB
}

func NewUserRepo(sqlConn *sql.DB) *UserRepo {
	return &UserRepo{
		SqlDB: sqlConn,
	}
}

func (r *UserRepo) CreateUser(username, host string, ssoId int, orgs []int) (*types.User, error) {
	v, ok := internal.Hosts[host]
	if !ok {
		return nil, fmt.Errorf("No such host: %s", host)
	}

	matchedOrgs := intersectOrgs(orgs, v.Orgs)
	if len(matchedOrgs) == 0 {
		return nil, fmt.Errorf("user %s has no orgs matching host %s", username, host)
	}

	user, err := r.UserExists(username, ssoId, v.Id)
	if err != nil {
		return nil, err
	}

	if user != nil {
		user.Orgs = matchedOrgs
		return user, nil
	}

	tx, err := r.SqlDB.Begin()
	if err != nil {
		return nil, err
	}

	var userID int
	err = tx.QueryRow(`INSERT INTO "user" (display_name, last_login, created_at) VALUES ($1, $2, $3) RETURNING id`,
		username,
		time.Now(),
		time.Now(),
	).Scan(&userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(`INSERT INTO auth (account, username, sso_id, host) VALUES ($1, $2, $3, $4)`,
		userID,
		username,
		ssoId,
		v.Id,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	user = &types.User{
		ID: userID,
		Username: username,
		DisplayName: username,
		SsoID: ssoId,
		Host: host,
		Orgs: matchedOrgs,
	}

	return user, nil
}

func intersectOrgs(userOrgs, hostOrgs []int) []int {
	allowed := make(map[int]struct{}, len(hostOrgs))
	for _, id := range hostOrgs {
		allowed[id] = struct{}{}
	}

	matched := make([]int, 0)
	for _, id := range userOrgs {
		if _, ok := allowed[id]; ok {
			matched = append(matched, id)
		}
	}
	return matched
}


func (r *UserRepo) SearchByName(hostID int, query string, excludeID int) ([]types.User, error) {
	rows, err := r.SqlDB.Query(
		`SELECT DISTINCT u.id, u.display_name, a.username, a.sso_id
		 FROM "user" u
		 JOIN "auth" a ON a.account = u.id
		 WHERE a.host = $1
		   AND u.display_name ILIKE $2
		   AND u.id <> $3
		 ORDER BY u.display_name ASC
		 LIMIT 25`,
		hostID, query+"%", excludeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]types.User, 0)
	for rows.Next() {
		var u types.User
		if err := rows.Scan(&u.ID, &u.DisplayName, &u.Username, &u.SsoID); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *UserRepo) UserExistsByID(id int) (bool, error) {
	var exists bool
	err := r.SqlDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM "user" WHERE id=$1)`, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *UserRepo) PurgeUserData(userID int) ([]string, error) {
	tx, err := r.SqlDB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.Query(
		`SELECT v.path
		 FROM "version" v
		 JOIN "product" pr ON pr.id = v.product
		 JOIN "project" p ON p.id = pr.project
		 WHERE p.owner = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			rows.Close()
			return nil, err
		}
		paths = append(paths, path)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	steps := []string{
		`DELETE FROM "version"
		 WHERE product IN (
			 SELECT pr.id FROM "product" pr
			 JOIN "project" p ON p.id = pr.project
			 WHERE p.owner = $1
		 )`,
		`DELETE FROM "product"
		 WHERE project IN (SELECT id FROM "project" WHERE owner = $1)`,
		`DELETE FROM "permission"
		 WHERE account = $1
		    OR project IN (SELECT id FROM "project" WHERE owner = $1)`,
		`DELETE FROM "project" WHERE owner = $1`,
		`DELETE FROM "token_access"
		 WHERE token IN (SELECT id FROM "token" WHERE owner = $1)`,
		`DELETE FROM "token" WHERE owner = $1`,
		`DELETE FROM "auth" WHERE account = $1`,
	}
	for _, q := range steps {
		if _, err := tx.Exec(q, userID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return paths, nil
}

func (r *UserRepo) UserExists(username string, ssoId, host int) (*types.User, error) {
	var user types.User
	err := r.SqlDB.QueryRow(
		`SELECT a.display_name, auth.account, auth.username, auth.sso_id, auth.host
		 FROM auth
		 JOIN "user" a ON a.id = auth.account
		 WHERE auth.username=$1 AND auth.sso_id=$2 AND auth.host=$3`,
		username,
		ssoId,
		host,
	).Scan(
		&user.DisplayName,
		&user.ID,
		&user.Username,
		&user.SsoID,
		&user.Host,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
