package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ValidateResponse struct {
	Valid bool `json:"valid"`
	Admin bool `json:"admin"`
}

func (h *AuthHandler) HandleValidate(c *gin.Context) {
	authHeader := c.GetHeader("X-Api-Token")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authorization header missing",
		})
		return
	}

	token, err := h.Repo.FetchToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ValidateResponse{
			Valid: false,
			Admin: false,
		})
		return
	}

	c.JSON(http.StatusOK, ValidateResponse{
		Valid: true,
		Admin: token.Admin,
	})
}
