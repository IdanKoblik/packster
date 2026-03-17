package product

import (
	"artifactor/internal/utils"
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

	if err := utils.ValidateName(request.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ignore any caller-supplied tokens; only grant the creating token full access.
	tokens := map[string]types.TokenPermissions{
		c.GetString("token"): {
			Download:   true,
			Upload:     true,
			Delete:     true,
			Maintainer: true,
		},
	}

	err = h.Repo.CreateProduct(&types.Product{
		Name:     request.Name,
		Tokens:   tokens,
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
