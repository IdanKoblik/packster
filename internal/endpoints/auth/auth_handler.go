package auth

import (
	"packster/internal/repository"
	"packster/pkg/types"
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
