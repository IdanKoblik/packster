package gitlab

import (
	"packster/internal/repository"
	"packster/pkg/config"
)

type GitlabHandler struct {
	Cfg config.Config
	Repo repository.IUserRepo
}

func NewGitlabHandler(cfg config.Config, repo repository.IUserRepo) *GitlabHandler {
	return &GitlabHandler{
		Cfg: cfg,
		Repo: repo,
	}
}
