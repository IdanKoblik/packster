package product

import (
	"net/http"
	"os"
	"packster/internal/endpoints"
	"packster/internal/utils"

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

	product := h.fetchProductOrAbort(c, productName, group)
	if product == nil {
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

	absFile, ok := validateFilePath(c, v.Path)
	if !ok {
		return
	}

	if err := os.Remove(absFile); err != nil && !os.IsNotExist(err) {
		endpoints.InternalError(c, err)
		return
	}

	if err := h.Repo.DeleteVersion(productName, group, version, c.GetString("token"), c.GetBool("admin")); err != nil {
		endpoints.InternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
