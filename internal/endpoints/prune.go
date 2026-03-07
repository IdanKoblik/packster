package endpoints

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) HandlePrune(c *gin.Context) {
	admin, exists := c.Get("admin")
	if !exists || !admin.(bool) {
		c.String(http.StatusUnauthorized, "Only admin allowed to prune an api token")
		return
	}

	token := c.Param("token")
	if token == "" {
		c.String(http.StatusBadRequest, "Missing api token")
		return
	}

	err := h.Repo.PruneToken(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.String(http.StatusNoContent, "Successfully deleted this api token")
}
