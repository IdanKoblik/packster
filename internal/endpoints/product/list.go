package product

import (
	"net/http"
	"packster/internal/utils"

	"github.com/gin-gonic/gin"
)

// HandleListProducts godoc
// @Summary      List products
// @Description  Returns all products for admins, or only products accessible to the token for non-admins.
// @Tags         products
// @Produce      json
// @Success      200  {array}   types.Product
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /product/list [get]
func (h *ProductHandler) HandleListProducts(c *gin.Context) {
	if c.GetBool("admin") {
		products, err := h.Repo.ListProducts()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, products)
	} else {
		hashedToken := utils.Hash(c.GetString("token"))
		products, err := h.Repo.ListProductsByToken(hashedToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, products)
	}
}
