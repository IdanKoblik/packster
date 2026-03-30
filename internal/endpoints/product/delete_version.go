package product

import (
	"net/http"
	"os"
	"packster/internal/utils"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// HandleDeleteVersion godoc
// @Summary      Delete a version artifact
// @Description  Removes the artifact file and version metadata for the specified product version. Requires Delete permission.
// @Tags         versions
// @Param        product  path   string  true  "Product name"
// @Param        version  path   string  true  "Version identifier"
// @Param        group    query  string  false "Product group (default: empty)"
// @Success      204  "Version deleted"
// @Failure      400  {string}  string  "Product or version not found"
// @Failure      403  {string}  string  "Permission denied"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /product/delete/{product}/{version} [delete]
func (h *ProductHandler) HandleDeleteVersion(c *gin.Context) {
	productName := c.Param("product")
	if productName == "" {
		c.String(http.StatusBadRequest, "product required")
		return
	}

	version := c.Param("version")
	if version == "" {
		c.String(http.StatusBadRequest, "version required")
		return
	}

	group := c.Query("group")

	product, err := h.Repo.FetchProduct(productName, group)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	if product == nil {
		c.String(http.StatusBadRequest, "Product not found")
		return
	}

	permissions := product.Tokens[utils.Hash(c.GetString("token"))]
	if !c.GetBool("admin") && !permissions.Delete {
		c.String(http.StatusForbidden, "permission denied")
		return
	}

	v, ok := product.Versions[version]
	if !ok {
		c.String(http.StatusBadRequest, "Version not found")
		return
	}

	// Defense-in-depth: ensure the stored path cannot escape the products directory.
	absBase, err := filepath.Abs(productsBaseDir)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}
	absFile, err := filepath.Abs(v.Path)
	if err != nil || !strings.HasPrefix(absFile, absBase+string(filepath.Separator)) {
		c.String(http.StatusForbidden, "invalid file path")
		return
	}

	if err := os.Remove(absFile); err != nil && !os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if err := h.Repo.DeleteVersion(productName, group, version, c.GetString("token"), c.GetBool("admin")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.Status(http.StatusNoContent)
}
