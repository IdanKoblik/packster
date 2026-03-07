package endpoints

import (
	"net/http"
	requests "artifactor/pkg/http"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) HandleRegister(c *gin.Context) {
	admin, exists := c.Get("admin")
	if !exists || !admin.(bool) {
		c.String(http.StatusUnauthorized, "Only admin allowed to register new tokens")
		return
	}

	var request requests.CreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.String(http.StatusBadRequest, "Missing request body")
		return
	}

	token, err := h.Repo.CreateToken(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.String(http.StatusCreated, token)
}
