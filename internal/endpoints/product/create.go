package product

import (
	"artifactor/pkg/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *ProductHandler) HandleCreate(c *gin.Context) {
	var request types.CreateProductRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	if request.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name is required",
		})

		return
	}

	if request.Tokens == nil {
		request.Tokens = make(map[string]types.TokenPermissions)
	}

	request.Tokens[c.GetString("token")] = types.TokenPermissions{
		Download:   true,
		Upload:     true,
		Delete:     true,
		Maintainer: true,
	}

	err = h.Repo.CreateProduct(&types.Product{
		Name:     request.Name,
		Tokens:   request.Tokens,
		Versions: map[string]types.Version{},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.Status(http.StatusCreated)
}
