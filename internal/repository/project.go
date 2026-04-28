package repository

import (
	"database/sql"
	"errors"
	"time"

	"packster/pkg/types"
)

type IProjectRepo interface {
	Import(ownerID, hostID, repository int) (*types.Project, error)
	GetByID(id int) (*types.Project, error)
	ListAccessible(userID int) ([]types.Project, error)
	GetByHostRepository(hostID, repository int) (*types.Project, error)
	Delete(id int) ([]string, error)
}

type ProjectRepo struct {
	SqlDB *sql.DB
}

func NewProjectRepo(sqlConn *sql.DB) *ProjectRepo {
	return &ProjectRepo{SqlDB: sqlConn}
}

func (r *ProjectRepo) Import(ownerID, hostID, repository int) (*types.Project, error) {
	tx, err := r.SqlDB.Begin()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var projectID int
	err = tx.QueryRow(
		`INSERT INTO "project" (host, repository, owner, created_at) VALUES ($1, $2, $3, $4) RETURNING id`,
		hostID, repository, ownerID, now,
	).Scan(&projectID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(
		`INSERT INTO "permission" (account, project, can_download, can_upload, can_delete) VALUES ($1, $2, TRUE, TRUE, TRUE)`,
		ownerID, projectID,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &types.Project{
		ID:         projectID,
		Host:       hostID,
		Repository: repository,
		Owner:      ownerID,
		CreatedAt:  now,
	}, nil
}

func (r *ProjectRepo) GetByID(id int) (*types.Project, error) {
	var p types.Project
	err := r.SqlDB.QueryRow(
		`SELECT id, host, repository, owner, created_at FROM "project" WHERE id=$1`,
		id,
	).Scan(&p.ID, &p.Host, &p.Repository, &p.Owner, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepo) GetByHostRepository(hostID, repository int) (*types.Project, error) {
	var p types.Project
	err := r.SqlDB.QueryRow(
		`SELECT id, host, repository, owner, created_at FROM "project" WHERE host=$1 AND repository=$2`,
		hostID, repository,
	).Scan(&p.ID, &p.Host, &p.Repository, &p.Owner, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepo) Delete(id int) ([]string, error) {
	tx, err := r.SqlDB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.Query(
		`SELECT v.path
		 FROM "version" v
		 JOIN "product" pr ON pr.id = v.product
		 WHERE pr.project = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			rows.Close()
			return nil, err
		}
		paths = append(paths, p)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	steps := []string{
		`DELETE FROM "version"
		 WHERE product IN (SELECT id FROM "product" WHERE project = $1)`,
		`DELETE FROM "token_access"
		 WHERE product IN (SELECT id FROM "product" WHERE project = $1)`,
		`DELETE FROM "product" WHERE project = $1`,
		`DELETE FROM "permission" WHERE project = $1`,
		`DELETE FROM "project" WHERE id = $1`,
	}
	for _, q := range steps {
		if _, err := tx.Exec(q, id); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return paths, nil
}

func (r *ProjectRepo) ListAccessible(userID int) ([]types.Project, error) {
	rows, err := r.SqlDB.Query(
		`SELECT p.id, p.host, p.repository, p.owner, p.created_at
		 FROM "project" p
		 JOIN "permission" pe ON pe.project = p.id
		 WHERE pe.account = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]types.Project, 0)
	for rows.Next() {
		var p types.Project
		if err := rows.Scan(&p.ID, &p.Host, &p.Repository, &p.Owner, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
