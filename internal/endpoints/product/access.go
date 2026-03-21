package product

import (
	"artifactor/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleAccess godoc
// @Summary      List accessible products
// @Description  Returns the names of all products the authenticated token has access to.
// @Tags         products
// @Produce      json
// @Success      200  {array}   string             "Product names accessible by the token"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /product/access [get]
func (h *ProductHandler) HandleAccess(c *gin.Context) {
	names, err := h.Repo.ListProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	token := utils.Hash(c.GetString("token"))
	var accessible []string

	for _, name := range names {
		product, err := h.Repo.FetchProduct(name)
		if err != nil {
			continue
		}

		if _, hasAccess := product.Tokens[token]; hasAccess {
			accessible = append(accessible, name)
		}
	}

	c.JSON(http.StatusOK, accessible)
}
