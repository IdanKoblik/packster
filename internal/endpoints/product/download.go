package product

import (
	"net/http"
	"packster/internal/metrics"
	"packster/internal/utils"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const productsBaseDir = "prodcuts"

// HandleDownload godoc
// @Summary      Download a version artifact
// @Description  Streams the artifact file for the specified product version. Requires Download permission.
// @Tags         versions
// @Produce      application/octet-stream
// @Param        product  path   string  true  "Product name"
// @Param        version  path   string  true  "Version identifier"
// @Param        group    query  string  false "Product group (default: empty)"
// @Success      200  {file}  binary  "Artifact file — Content-Disposition: attachment; filename=\"<original filename>\""
// @Failure      400  {string}  string  "Product or version not found"
// @Failure      403  {string}  string  "Permission denied"
// @Failure      500  {string}  string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /product/download/{product}/{version} [get]
func (h *ProductHandler) HandleDownload(c *gin.Context) {
	productName := c.Param("product")
	version := c.Param("version")
	group := c.Query("group")

	product := h.fetchProductOrAbort(c, productName, group)
	if product == nil {
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

	absFile, ok := validateFilePath(c, v.Path)
	if !ok {
		return
	}

	metrics.ArtifactDownloadsTotal.WithLabelValues(productName).Inc()
	c.FileAttachment(absFile, filepath.Base(absFile))
}
