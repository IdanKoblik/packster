package product

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleDelete godoc
// @Summary      Delete a product
// @Description  Permanently removes a product and all its metadata. Requires maintainer or admin access.
// @Tags         products
// @Param        product  path   string  true  "Product name"
// @Param        group    query  string  false "Product group (default: empty)"
// @Success      204  "Product deleted"
// @Failure      400  {string}  string  "Missing product name"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /product/delete/{product} [delete]
func (h *ProductHandler) HandleDelete(c *gin.Context) {
	product := c.Param("product")
	if product == "" {
		c.String(http.StatusBadRequest, "product required")
		return
	}

	group := c.Query("group")

	err := h.Repo.DeleteProduct(product, group, c.GetString("token"), c.GetBool("admin"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.Status(http.StatusNoContent)
}
