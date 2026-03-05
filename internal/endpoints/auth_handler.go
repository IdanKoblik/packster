package endpoints

import (
	"artifactor/internal/repository"
	"artifactor/pkg/requests"
)

type AuthRepo interface {
	UserExists(username string) (bool, error)
	CreateUser(request *requests.RegisterRequest) error
}

type AuthHandler struct {
	Repo AuthRepo
}

func NewAuthHandler(repo *repository.AuthRepository) *AuthHandler {
	return &AuthHandler{Repo: repo}
}
