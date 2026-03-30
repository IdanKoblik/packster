package product

import (
	"net/http"
	"packster/internal/utils"

	"github.com/gin-gonic/gin"
)

// HandleFetch godoc
// @Summary      Fetch a product
// @Description  Returns full product metadata including tokens and versions. Requires token access or admin privileges.
// @Tags         products
// @Produce      json
// @Param        product  path   string  true  "Product name"
// @Param        group    query  string  false "Product group (default: empty)"
// @Success      200  {object}  types.Product  "Product metadata"
// @Failure      400  {string}  string  "Missing product name"
// @Failure      403  {string}  string  "Permission denied"
// @Failure      404  {string}  string  "Product not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /product/fetch/{product} [get]
func (h *ProductHandler) HandleFetch(c *gin.Context) {
	productName := c.Param("product")
	if productName == "" {
		c.String(http.StatusBadRequest, "product required")
		return
	}

	group := c.Query("group")

	product, err := h.Repo.FetchProduct(productName, group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if product == nil {
		c.String(http.StatusNotFound, "Product not found")
		return
	}

	_, hasAccess := product.Tokens[utils.Hash(c.GetString("token"))]
	if !c.GetBool("admin") && !hasAccess {
		c.String(http.StatusForbidden, "permission denied")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":       product.Name,
		"group_name": product.GroupName,
		"versions":   product.Versions,
	})
}
