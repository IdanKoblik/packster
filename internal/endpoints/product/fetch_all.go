package product

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleFetchAll godoc
// @Summary      Fetch all products
// @Description  Returns all product metadata. Requires admin privileges.
// @Tags         products
// @Produce      json
// @Success      200  {array}  types.Product  "List of all products"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /product/fetch [get]
func (h *ProductHandler) HandleFetchAll(c *gin.Context) {
	if !c.GetBool("admin") {
		c.String(http.StatusForbidden, "admin access required")
		return
	}

	products, err := h.Repo.FetchAllProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, products)
}
