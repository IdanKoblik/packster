package product

import (
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

	c.JSON(http.StatusOK, product)
}
