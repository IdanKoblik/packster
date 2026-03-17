package product

import (
	"artifactor/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *ProductHandler) HandleFetch(c *gin.Context) {
	productName := c.Param("product")
	if productName == "" {
		c.String(http.StatusBadRequest, "product required")
		return
	}

	product, err := h.Repo.FetchProduct(productName)
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

	c.JSON(http.StatusOK, product)
}
