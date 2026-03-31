package auth

import (
	"net/http"
	"packster/internal/endpoints"

	"github.com/gin-gonic/gin"
)

// HandlePrune godoc
// @Summary      Delete an API token
// @Description  Permanently removes the specified API token. Requires admin privileges.
// @Tags         auth
// @Param        token  path  string  true  "Token to delete"
// @Success      204  "Token deleted"
// @Failure      400  {string}  string  "Missing token"
// @Failure      401  {string}  string  "Admin privileges required"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /prune/{token} [delete]
func (h *AuthHandler) HandlePrune(c *gin.Context) {
	if !h.requireAdmin(c, "prune an api token") {
		return
	}

	token := c.Param("token")
	if token == "" {
		c.String(http.StatusBadRequest, "Missing api token")
		return
	}

	err := h.Repo.PruneToken(token)
	if err != nil {
		endpoints.InternalError(c, err)
	}

	c.String(http.StatusNoContent, "Successfully deleted this api token")
}
