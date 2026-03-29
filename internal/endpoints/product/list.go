package product

import (
	"packster/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleListProducts godoc
// @Summary      List products
// @Description  Returns all product names for admins, or only products accessible to the token for non-admins.
// @Tags         products
// @Produce      json
// @Success      200  {array}   string
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /product/list [get]
func (h *ProductHandler) HandleListProducts(c *gin.Context) {
	var (
		names []string
		err   error
	)

	if c.GetBool("admin") {
		names, err = h.Repo.ListProducts()
	} else {
		hashedToken := utils.Hash(c.GetString("token"))
		names, err = h.Repo.ListProductsByToken(hashedToken)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, names)
}
