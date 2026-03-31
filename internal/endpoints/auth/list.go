package auth

import (
	"net/http"
	"packster/internal/endpoints"

	"github.com/gin-gonic/gin"
)

// HandleListTokens godoc
// @Summary      List all API tokens
// @Description  Returns all registered API tokens. Requires admin privileges.
// @Tags         auth
// @Produce      json
// @Success      200  {array}   types.ApiToken
// @Failure      401  {string}  string             "Admin privileges required"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /tokens [get]
func (h *AuthHandler) HandleListTokens(c *gin.Context) {
	if !h.requireAdmin(c, "list tokens") {
		return
	}

	tokens, err := h.Repo.ListTokens()
	if err != nil {
		endpoints.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, tokens)
}
