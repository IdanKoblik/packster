package projects

import (
	"net/http"

	"packster/internal/repository"
	"packster/pkg/config"
)

type ProjectsHandler struct {
	Cfg            config.Config
	UserRepo       repository.IUserRepo
	ProjectRepo    repository.IProjectRepo
	PermissionRepo repository.IPermissionRepo
	ProductRepo    repository.IProductRepo
	VersionRepo    repository.IVersionRepo
	HTTP           *http.Client
}

func NewProjectsHandler(
	cfg config.Config,
	userRepo repository.IUserRepo,
	projectRepo repository.IProjectRepo,
	permRepo repository.IPermissionRepo,
	productRepo repository.IProductRepo,
	versionRepo repository.IVersionRepo,
) *ProjectsHandler {
	return &ProjectsHandler{
		Cfg:            cfg,
		UserRepo:       userRepo,
		ProjectRepo:    projectRepo,
		PermissionRepo: permRepo,
		ProductRepo:    productRepo,
		VersionRepo:    versionRepo,
		HTTP:           &http.Client{},
	}
}
