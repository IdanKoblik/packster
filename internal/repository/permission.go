package repository

import (
	"database/sql"
	"errors"

	"packster/pkg/types"
)

type IPermissionRepo interface {
	Get(userID, projectID int) (*types.Permission, error)
	Set(perm types.Permission) error
	Delete(userID, projectID int) error
	ListByProject(projectID int) ([]types.PermissionEntry, error)
}

type PermissionRepo struct {
	SqlDB *sql.DB
}

func NewPermissionRepo(sqlConn *sql.DB) *PermissionRepo {
	return &PermissionRepo{SqlDB: sqlConn}
}

func (r *PermissionRepo) Get(userID, projectID int) (*types.Permission, error) {
	var p types.Permission
	err := r.SqlDB.QueryRow(
		`SELECT account, project, can_download, can_upload, can_delete
		 FROM "permission" WHERE account=$1 AND project=$2`,
		userID, projectID,
	).Scan(&p.Account, &p.Project, &p.CanDownload, &p.CanUpload, &p.CanDelete)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PermissionRepo) Set(perm types.Permission) error {
	_, err := r.SqlDB.Exec(
		`INSERT INTO "permission" (account, project, can_download, can_upload, can_delete)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (account, project) DO UPDATE SET
		   can_download = EXCLUDED.can_download,
		   can_upload   = EXCLUDED.can_upload,
		   can_delete   = EXCLUDED.can_delete`,
		perm.Account, perm.Project, perm.CanDownload, perm.CanUpload, perm.CanDelete,
	)
	return err
}

func (r *PermissionRepo) ListByProject(projectID int) ([]types.PermissionEntry, error) {
	rows, err := r.SqlDB.Query(
		`SELECT pe.account, pe.project, pe.can_download, pe.can_upload, pe.can_delete, u.display_name
		 FROM "permission" pe
		 JOIN "user" u ON u.id = pe.account
		 WHERE pe.project = $1
		 ORDER BY u.display_name ASC`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]types.PermissionEntry, 0)
	for rows.Next() {
		var e types.PermissionEntry
		if err := rows.Scan(&e.Account, &e.Project, &e.CanDownload, &e.CanUpload, &e.CanDelete, &e.DisplayName); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *PermissionRepo) Delete(userID, projectID int) error {
	_, err := r.SqlDB.Exec(
		`DELETE FROM "permission" WHERE account=$1 AND project=$2`,
		userID, projectID,
	)
	return err
}
