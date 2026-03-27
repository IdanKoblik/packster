package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleListTokens godoc
// @Summary      List all API tokens
// @Description  Returns the hashed IDs of all registered tokens. Requires admin privileges.
// @Tags         auth
// @Produce      json
// @Success      200  {array}   string
// @Failure      401  {string}  string             "Admin privileges required"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /tokens [get]
func (h *AuthHandler) HandleListTokens(c *gin.Context) {
	admin, exists := c.Get("admin")
	if !exists || !admin.(bool) {
		c.String(http.StatusUnauthorized, "Only admin allowed to list tokens")
		return
	}

	tokens, err := h.Repo.ListTokens()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tokens)
}
