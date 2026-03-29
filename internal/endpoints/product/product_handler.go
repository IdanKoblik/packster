package product

import (
	"packster/internal/repository"
	"packster/pkg/types"
)

type ProductRepo interface {
	CreateProduct(product *types.Product) error
	DeleteProduct(name, token string, admin bool) error
	FetchProduct(name string) (*types.Product, error)
	DeleteToken(productName, sourceToken, targetToken string, admin bool) error
	AddToken(productName, sourceToken, targetToken string, permissions types.TokenPermissions, admin bool) error
	AddVersion(productName, version, token string, admin bool, v types.Version) error
	DeleteVersion(productName, version, token string, admin bool) error
	ListProducts() ([]string, error)
	ListProductsByToken(hashedToken string) ([]string, error)
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
