package product

import (
	"net/http"
	"packster/internal/utils"
	"packster/pkg/types"

	"github.com/gin-gonic/gin"
)

// HandleAccess godoc
// @Summary      List accessible products
// @Description  Returns the name and group of all products the authenticated token has access to.
// @Tags         products
// @Produce      json
// @Success      200  {array}   types.Product      "Products accessible by the token"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /product/access [get]
func (h *ProductHandler) HandleAccess(c *gin.Context) {
	products, err := h.Repo.ListProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	token := utils.Hash(c.GetString("token"))
	var accessible []types.Product

	for _, p := range products {
		product, err := h.Repo.FetchProduct(p.Name, p.GroupName)
		if err != nil {
			continue
		}

		if _, hasAccess := product.Tokens[token]; hasAccess {
			accessible = append(accessible, types.Product{Name: p.Name, GroupName: p.GroupName})
		}
	}

	c.JSON(http.StatusOK, accessible)
}
