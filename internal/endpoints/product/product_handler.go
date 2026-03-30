package product

import (
	"packster/internal/repository"
	"packster/pkg/types"
)

type ProductRepo interface {
	CreateProduct(product *types.Product) error
	DeleteProduct(name, group, token string, admin bool) error
	FetchProduct(name, group string) (*types.Product, error)
	DeleteToken(productName, group, sourceToken, targetToken string, admin bool) error
	AddToken(productName, group, sourceToken, targetToken string, permissions types.TokenPermissions, admin bool) error
	AddVersion(productName, group, version, token string, admin bool, v types.Version) error
	DeleteVersion(productName, group, version, token string, admin bool) error
	ListProducts() ([]types.Product, error)
	ListProductsByToken(hashedToken string) ([]types.Product, error)
}

type ProductHandler struct {
	Repo          ProductRepo
	FileSizeLimit int
}

func NewProductHandler(repo *repository.ProductRepository, fileSizeLimit int) *ProductHandler {
	return &ProductHandler{
		Repo:          repo,
		FileSizeLimit: fileSizeLimit,
	}
}
