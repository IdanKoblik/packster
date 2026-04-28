package repository

import (
	"database/sql"
	"errors"
	"time"

	"packster/pkg/types"
)

type IProductRepo interface {
	Create(projectID int, name string) (*types.Product, error)
	GetByID(id int) (*types.Product, error)
	GetByName(projectID int, name string) (*types.Product, error)
	ListByProject(projectID int) ([]types.Product, error)
	Delete(id int) error
}

type ProductRepo struct {
	SqlDB *sql.DB
}

func NewProductRepo(sqlConn *sql.DB) *ProductRepo {
	return &ProductRepo{SqlDB: sqlConn}
}

func (r *ProductRepo) Create(projectID int, name string) (*types.Product, error) {
	now := time.Now()
	var id int
	err := r.SqlDB.QueryRow(
		`INSERT INTO "product" (name, project, created_at) VALUES ($1, $2, $3) RETURNING id`,
		name, projectID, now,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &types.Product{ID: id, Name: name, Project: projectID, CreatedAt: now}, nil
}

func (r *ProductRepo) GetByName(projectID int, name string) (*types.Product, error) {
	var p types.Product
	err := r.SqlDB.QueryRow(
		`SELECT id, name, project, created_at FROM "product" WHERE project=$1 AND name=$2`,
		projectID, name,
	).Scan(&p.ID, &p.Name, &p.Project, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) GetByID(id int) (*types.Product, error) {
	var p types.Product
	err := r.SqlDB.QueryRow(
		`SELECT id, name, project, created_at FROM "product" WHERE id=$1`,
		id,
	).Scan(&p.ID, &p.Name, &p.Project, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) ListByProject(projectID int) ([]types.Product, error) {
	rows, err := r.SqlDB.Query(
		`SELECT id, name, project, created_at FROM "product" WHERE project=$1 ORDER BY created_at ASC`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]types.Product, 0)
	for rows.Next() {
		var p types.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Project, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProductRepo) Delete(id int) error {
	_, err := r.SqlDB.Exec(`DELETE FROM "product" WHERE id=$1`, id)
	return err
}
