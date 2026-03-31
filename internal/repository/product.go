package repository

import (
	"context"
	"database/sql"
	"errors"
	"packster/internal/utils"
	"packster/pkg/config"
	"packster/pkg/types"
	"time"
)

type IProductRepo interface {
	CreateProduct(product *types.Product) error
	DeleteProduct(name, group, token string, admin bool) error
	FetchProduct(name, group string) (*types.Product, error)
	DeleteToken(productName, group, sourceToken, targetToken string, admin bool) error
	AddToken(productName, group, sourceToken, targetToken string, permissions types.TokenPermissions, admin bool) error
	DeleteVersion(productName, group, version, token string, admin bool) error
	AddVersion(productName, group, version, token string, admin bool, v types.Version) error
	ListProducts() ([]types.Product, error)
	ListProductsByToken(hashedToken string) ([]types.Product, error)
}

type ProductRepository struct {
	DB  *sql.DB
	Cfg *config.Config
}

func NewProductRepository(db *sql.DB, cfg *config.Config) *ProductRepository {
	return &ProductRepository{
		DB:  db,
		Cfg: cfg,
	}
}

func dbCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (r *ProductRepository) fetchAndAuthorize(name, group, token string, admin bool, check func(types.TokenPermissions) bool, errMsg string) (*types.Product, error) {
	product, err := r.FetchProduct(name, group)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	if !admin && !check(product.Tokens[utils.Hash(token)]) {
		return nil, errors.New(errMsg)
	}
	return product, nil
}

func scanProduct(rows *sql.Rows) (types.Product, error) {
	var p types.Product
	if err := rows.Scan(&p.Name, &p.GroupName); err != nil {
		return p, err
	}
	p.Tokens = map[string]types.TokenPermissions{}
	p.Versions = map[string]types.Version{}
	return p, nil
}

func (r *ProductRepository) ListProducts() ([]types.Product, error) {
	ctx, cancel := dbCtx()
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, "SELECT name, group_name FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []types.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *ProductRepository) ListProductsByToken(hashedToken string) ([]types.Product, error) {
	ctx, cancel := dbCtx()
	defer cancel()

	rows, err := r.DB.QueryContext(ctx,
		`SELECT p.name, p.group_name
		 FROM products p
		 JOIN product_permissions pp ON p.id = pp.product_id
		 JOIN api_tokens at ON pp.principal_id = at.id
		 WHERE at.token_hash = ?`, hashedToken)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []types.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *ProductRepository) FetchProduct(name, group string) (*types.Product, error) {
	ctx, cancel := dbCtx()
	defer cancel()

	var productID int64
	err := r.DB.QueryRowContext(ctx,
		"SELECT id FROM products WHERE name = ? AND group_name = ?", name, group).Scan(&productID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	permRows, err := r.DB.QueryContext(ctx,
		`SELECT at.token_hash, pp.can_download, pp.can_upload, pp.can_remove, pp.is_maintainer
		 FROM product_permissions pp
		 JOIN api_tokens at ON pp.principal_id = at.id
		 WHERE pp.product_id = ?`, productID)
	if err != nil {
		return nil, err
	}
	defer permRows.Close()

	tokens := make(map[string]types.TokenPermissions)
	for permRows.Next() {
		var tokenHash string
		var perms types.TokenPermissions
		if err := permRows.Scan(&tokenHash, &perms.Download, &perms.Upload, &perms.Delete, &perms.Maintainer); err != nil {
			return nil, err
		}
		tokens[tokenHash] = perms
	}
	permRows.Close()

	verRows, err := r.DB.QueryContext(ctx,
		"SELECT name, path, checksum FROM product_versions WHERE product_id = ?", productID)
	if err != nil {
		return nil, err
	}
	defer verRows.Close()

	versions := make(map[string]types.Version)
	for verRows.Next() {
		var vName, vPath, vChecksum string
		if err := verRows.Scan(&vName, &vPath, &vChecksum); err != nil {
			return nil, err
		}
		versions[vName] = types.Version{Path: vPath, Checksum: vChecksum}
	}

	return &types.Product{
		Name:      name,
		GroupName: group,
		Tokens:    tokens,
		Versions:  versions,
	}, nil
}

func (r *ProductRepository) CreateProduct(product *types.Product) error {
	existing, err := r.FetchProduct(product.Name, product.GroupName)
	if err != nil {
		return err
	}

	if existing != nil {
		return errors.New("product already exists")
	}

	product.HashTokens()

	ctx, cancel := dbCtx()
	defer cancel()

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		"INSERT INTO products (name, group_name) VALUES (?, ?)", product.Name, product.GroupName)
	if err != nil {
		return err
	}

	productID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	for hashedToken, perms := range product.Tokens {
		var principalID int64
		err := tx.QueryRowContext(ctx, "SELECT id FROM api_tokens WHERE token_hash = ?", hashedToken).Scan(&principalID)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx,
			`INSERT INTO product_permissions (principal_id, product_id, can_download, can_upload, can_remove, is_maintainer)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			principalID, productID, perms.Download, perms.Upload, perms.Delete, perms.Maintainer)
		if err != nil {
			return err
		}
	}

	for name, v := range product.Versions {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO product_versions (product_id, name, path, checksum) VALUES (?, ?, ?, ?)",
			productID, name, v.Path, v.Checksum)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ProductRepository) DeleteProduct(name, group, token string, admin bool) error {
	if _, err := r.fetchAndAuthorize(name, group, token, admin, func(p types.TokenPermissions) bool {
		return p.Maintainer || p.Delete
	}, "missing delete permission"); err != nil {
		return err
	}

	ctx, cancel := dbCtx()
	defer cancel()

	_, err := r.DB.ExecContext(ctx, "DELETE FROM products WHERE name = ? AND group_name = ?", name, group)
	return err
}

func (r *ProductRepository) DeleteToken(productName, group, sourceToken, targetToken string, admin bool) error {
	if _, err := r.fetchAndAuthorize(productName, group, sourceToken, admin, func(p types.TokenPermissions) bool {
		return p.Maintainer
	}, "missing maintainer permission"); err != nil {
		return err
	}

	ctx, cancel := dbCtx()
	defer cancel()

	_, err := r.DB.ExecContext(ctx,
		`DELETE pp FROM product_permissions pp
		 JOIN products p ON pp.product_id = p.id
		 JOIN api_tokens at ON pp.principal_id = at.id
		 WHERE p.name = ? AND p.group_name = ? AND at.token_hash = ?`,
		productName, group, utils.Hash(targetToken))
	return err
}

func (r *ProductRepository) AddToken(productName, group, sourceToken, targetToken string, permissions types.TokenPermissions, admin bool) error {
	if _, err := r.fetchAndAuthorize(productName, group, sourceToken, admin, func(p types.TokenPermissions) bool {
		return p.Maintainer
	}, "missing maintainer permission"); err != nil {
		return err
	}

	ctx, cancel := dbCtx()
	defer cancel()

	var productID int64
	err := r.DB.QueryRowContext(ctx,
		"SELECT id FROM products WHERE name = ? AND group_name = ?", productName, group).Scan(&productID)
	if err != nil {
		return err
	}

	var principalID int64
	err = r.DB.QueryRowContext(ctx, "SELECT id FROM api_tokens WHERE token_hash = ?", utils.Hash(targetToken)).Scan(&principalID)
	if err != nil {
		return err
	}

	_, err = r.DB.ExecContext(ctx,
		`INSERT INTO product_permissions (principal_id, product_id, can_download, can_upload, can_remove, is_maintainer)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		 can_download = VALUES(can_download), can_upload = VALUES(can_upload),
		 can_remove = VALUES(can_remove), is_maintainer = VALUES(is_maintainer)`,
		principalID, productID, permissions.Download, permissions.Upload, permissions.Delete, permissions.Maintainer)
	return err
}

func (r *ProductRepository) DeleteVersion(productName, group, version, token string, admin bool) error {
	if _, err := r.fetchAndAuthorize(productName, group, token, admin, func(p types.TokenPermissions) bool {
		return p.Maintainer || p.Delete
	}, "missing maintainer / delete permission"); err != nil {
		return err
	}

	ctx, cancel := dbCtx()
	defer cancel()

	_, err := r.DB.ExecContext(ctx,
		`DELETE pv FROM product_versions pv
		 JOIN products p ON pv.product_id = p.id
		 WHERE p.name = ? AND p.group_name = ? AND pv.name = ?`,
		productName, group, version)
	return err
}

func (r *ProductRepository) AddVersion(productName, group, version, token string, admin bool, v types.Version) error {
	product, err := r.fetchAndAuthorize(productName, group, token, admin, func(p types.TokenPermissions) bool {
		return p.Upload
	}, "missing upload permission")
	if err != nil {
		return err
	}

	if _, ok := product.Versions[version]; ok {
		return errors.New("version already exists")
	}

	ctx, cancel := dbCtx()
	defer cancel()

	var productID int64
	err = r.DB.QueryRowContext(ctx,
		"SELECT id FROM products WHERE name = ? AND group_name = ?", productName, group).Scan(&productID)
	if err != nil {
		return err
	}

	_, err = r.DB.ExecContext(ctx,
		"INSERT INTO product_versions (product_id, name, path, checksum) VALUES (?, ?, ?, ?)",
		productID, version, v.Path, v.Checksum)
	return err
}
