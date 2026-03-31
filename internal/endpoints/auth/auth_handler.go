package auth

import (
	"net/http"
	"packster/internal/repository"
	"packster/pkg/types"

	"github.com/gin-gonic/gin"
)

type AuthRepo interface {
	TokenExists(token string) (bool, error)
	CreateToken(request *types.RegisterRequest) (string, error)
	PruneToken(token string) error
	IsAdmin(token string) (bool, error)
	FetchToken(token string) (*types.ApiToken, error)
	ListTokens() ([]types.ApiToken, error)
}

type AuthHandler struct {
	Repo AuthRepo
}

func NewAuthHandler(repo *repository.AuthRepository) *AuthHandler {
	return &AuthHandler{Repo: repo}
}

// requireAdmin writes 401 and returns false if the caller is not an admin.
func (h *AuthHandler) requireAdmin(c *gin.Context, action string) bool {
	admin, exists := c.Get("admin")
	if !exists || !admin.(bool) {
		c.String(http.StatusUnauthorized, "Only admin allowed to "+action)
		return false
	}
	return true
}
