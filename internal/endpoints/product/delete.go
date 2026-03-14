package product

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *ProductHandler) HandleDelete(c *gin.Context) {
	product := c.Param("product")
	if product == "" {
		c.String(http.StatusBadRequest, "product required")
		return
	}

	err := h.Repo.DeleteProduct(product, c.GetString("token"), c.GetBool("admin"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.Status(http.StatusNoContent)
}
