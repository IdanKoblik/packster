package endpoints

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) HandleFetch(c *gin.Context) {
	admin, exists := c.Get("admin")
	if !exists || !admin.(bool) {
		c.String(http.StatusUnauthorized, "Only admin allowed to prune an api token")
		return
	}

	rawToken := c.Param("token")
	if rawToken == "" {
		c.String(http.StatusBadRequest, "Missing api token")
		return
	}

	token, err := h.Repo.FetchToken(rawToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, token)
}
