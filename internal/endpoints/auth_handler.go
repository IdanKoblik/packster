package endpoints

import (
	"artifactor/internal/repository"
	"artifactor/pkg/http"
	"artifactor/pkg/tokens"
)

type AuthRepo interface {
	TokenExists(token string) (bool, error)
	CreateToken(request *http.RegisterRequest) (string, error)
	PruneToken(token string) error
	IsAdmin(token string) (bool, error)
	FetchToken(token string) (*tokens.ApiToken, error)
}

type AuthHandler struct {
	Repo AuthRepo
}

func NewAuthHandler(repo *repository.AuthRepository) *AuthHandler {
	return &AuthHandler{Repo: repo}
}
