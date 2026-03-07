package endpoints

import (
	"artifactor/internal/repository"
	"artifactor/pkg/http"
	"artifactor/pkg/tokens"
)

type AuthRepo interface {
	FetchToken(token string) (*tokens.Token, error)
	CreateToken(request *http.CreateRequest) (string, error)
}

type AuthHandler struct {
	Repo AuthRepo
}

func NewAuthHandler(repo *repository.AuthRepository) *AuthHandler {
	return &AuthHandler{Repo: repo}
}
