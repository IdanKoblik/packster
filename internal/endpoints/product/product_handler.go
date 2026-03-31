package product

import (
	"net/http"
	"packster/internal/endpoints"
	"packster/internal/repository"
	"packster/pkg/types"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
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

// fetchProductOrAbort fetches a product; writes the response and returns nil on error or not-found.
func (h *ProductHandler) fetchProductOrAbort(c *gin.Context, name, group string) *types.Product {
	product, err := h.Repo.FetchProduct(name, group)
	if err != nil {
		endpoints.BadRequest(c, err)
		return nil
	}
	if product == nil {
		c.String(http.StatusBadRequest, "Product not found")
		return nil
	}
	return product
}

// validateFilePath ensures path is within productsBaseDir.
// Returns the absolute path and true on success, or writes the response and returns ("", false).
func validateFilePath(c *gin.Context, path string) (string, bool) {
	absBase, err := filepath.Abs(productsBaseDir)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return "", false
	}
	absFile, err := filepath.Abs(path)
	if err != nil || !strings.HasPrefix(absFile, absBase+string(filepath.Separator)) {
		c.String(http.StatusForbidden, "invalid file path")
		return "", false
	}
	return absFile, true
}
