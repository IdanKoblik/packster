package product

import (
	"packster/internal/metrics"
	"packster/internal/utils"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const productsBaseDir = "prodcuts"

// HandleDownload godoc
// @Summary      Download a version artifact
// @Description  Streams the artifact file for the specified product version. Requires Download permission.
// @Tags         versions
// @Produce      application/octet-stream
// @Param        product  path  string  true  "Product name"
// @Param        version  path  string  true  "Version identifier"
// @Success      200  {file}  binary  "Artifact file — Content-Disposition: attachment; filename=\"<original filename>\""
// @Failure      400  {string}  string  "Product or version not found"
// @Failure      403  {string}  string  "Permission denied"
// @Failure      500  {string}  string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /product/download/{product}/{version} [get]
func (h *ProductHandler) HandleDownload(c *gin.Context) {
	productName := c.Param("product")
	version := c.Param("version")

	product, err := h.Repo.FetchProduct(productName)
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
	if !c.GetBool("admin") && !permissions.Download {
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

	metrics.ArtifactDownloadsTotal.WithLabelValues(productName).Inc()
	c.FileAttachment(v.Path, filepath.Base(v.Path))
}
