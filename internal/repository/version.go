package repository

import (
	"database/sql"
	"errors"

	"packster/pkg/types"
)

type IVersionRepo interface {
	Create(productID int, name, path, checksum string) (*types.Version, error)
	GetByID(id int) (*types.Version, error)
	GetByName(productID int, name string) (*types.Version, error)
	ListByProduct(productID int) ([]types.Version, error)
	Delete(id int) error
}

type VersionRepo struct {
	SqlDB *sql.DB
}

func NewVersionRepo(sqlConn *sql.DB) *VersionRepo {
	return &VersionRepo{SqlDB: sqlConn}
}

func (r *VersionRepo) Create(productID int, name, path, checksum string) (*types.Version, error) {
	var id int
	err := r.SqlDB.QueryRow(
		`INSERT INTO "version" (name, path, checksum, product) VALUES ($1, $2, $3, $4) RETURNING id`,
		name, path, checksum, productID,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &types.Version{ID: id, Name: name, Path: path, Checksum: checksum, Product: productID}, nil
}

func (r *VersionRepo) GetByID(id int) (*types.Version, error) {
	var v types.Version
	err := r.SqlDB.QueryRow(
		`SELECT id, name, path, checksum, product FROM "version" WHERE id=$1`,
		id,
	).Scan(&v.ID, &v.Name, &v.Path, &v.Checksum, &v.Product)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VersionRepo) GetByName(productID int, name string) (*types.Version, error) {
	var v types.Version
	err := r.SqlDB.QueryRow(
		`SELECT id, name, path, checksum, product FROM "version" WHERE product=$1 AND name=$2`,
		productID, name,
	).Scan(&v.ID, &v.Name, &v.Path, &v.Checksum, &v.Product)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VersionRepo) ListByProduct(productID int) ([]types.Version, error) {
	rows, err := r.SqlDB.Query(
		`SELECT id, name, path, checksum, product FROM "version" WHERE product=$1 ORDER BY id ASC`,
		productID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]types.Version, 0)
	for rows.Next() {
		var v types.Version
		if err := rows.Scan(&v.ID, &v.Name, &v.Path, &v.Checksum, &v.Product); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *VersionRepo) Delete(id int) error {
	_, err := r.SqlDB.Exec(`DELETE FROM "version" WHERE id=$1`, id)
	return err
}
