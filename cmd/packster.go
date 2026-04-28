package main

import (
	"fmt"
	"os"

	"packster/internal"
	internalConfig "packster/internal/config"
	"packster/internal/endpoints"
	"packster/internal/endpoints/gitlab"
	"packster/internal/endpoints/projects"
	"packster/internal/logging"
	"packster/internal/repository"
	"packster/internal/sql"
	"packster/internal/ui"
	"packster/pkg/config"

	"github.com/gin-gonic/gin"
)

const defaultMultipartMemory = 32 << 20

const BANNER = `
██████╗  █████╗  ██████╗██╗  ██╗███████╗████████╗███████╗██████╗
██╔══██╗██╔══██╗██╔════╝██║ ██╔╝██╔════╝╚══██╔══╝██╔════╝██╔══██╗
██████╔╝███████║██║     █████╔╝ ███████╗   ██║   █████╗  ██████╔╝
██╔═══╝ ██╔══██║██║     ██╔═██╗ ╚════██║   ██║   ██╔══╝  ██╔══██╗
██║     ██║  ██║╚██████╗██║  ██╗███████║   ██║   ███████╗██║  ██║
╚═╝     ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚══════╝   ╚═╝   ╚══════╝╚═╝  ╚═╝
`

const (
	MAINTAINER = "Idan Koblik"

	PURPLE = "\033[38;2;87;87;232m"
	RESET = "\033[0m"
)

var BUILD_TIME string

func main() {
	logging.SetupLogger()
	printBanner()

	cfg, err := internalConfig.ParseConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		logging.Log.Error(err)
		os.Exit(1)
	}

	logging.Log.Debugf("Max file size that can be uploaded: %d MB\n", cfg.FileUploadLimit)

	_, err = sql.OpenPgsqlConnection(&cfg.Sql)
	if err != nil {
		logging.Log.Error(err)
		os.Exit(1)
	}

	defer sql.PgsqlConn.Close()
	logging.Log.Info("Successfully connected to pgsql db")

	logging.Log.Info("Loading hosts")
	err = internal.FetchHosts(cfg, sql.PgsqlConn)
	if err != nil {
		logging.Log.Error(err)
		os.Exit(1)
	}

	logging.Log.Info("Successfully loaded hosts")

	if cfg.Storage.Path != "" {
		if err := os.MkdirAll(cfg.Storage.Path, 0o755); err != nil {
			logging.Log.Errorf("failed to create storage path %q: %v", cfg.Storage.Path, err)
			os.Exit(1)
		}
	}

	logging.Log.Info("Setting up rest api")
	router := gin.Default()
	router.MaxMultipartMemory = defaultMultipartMemory

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "0.0.0.0:8080"
	}

	logging.Log.Debugf("Addr: %s", addr)

	api := router.Group("/api")
	{
		api.GET("/health", func(c *gin.Context){
			endpoints.HandleHealth(c, sql.PgsqlConn)
		})

		api.GET("/hosts", func(c *gin.Context){
			endpoints.HandleHosts(c, internal.Hosts)
		})
	}

	userRepo := repository.NewUserRepo(sql.PgsqlConn)
	projectRepo := repository.NewProjectRepo(sql.PgsqlConn)
	permRepo := repository.NewPermissionRepo(sql.PgsqlConn)
	productRepo := repository.NewProductRepo(sql.PgsqlConn)
	versionRepo := repository.NewVersionRepo(sql.PgsqlConn)

	registerGitalbEndpoints(cfg, api, userRepo)
	registerProjectEndpoints(cfg, api, userRepo, projectRepo, permRepo, productRepo, versionRepo)
	ui.RegisterRoutes(router)

	err = router.Run(addr)
	if err != nil {
		logging.Log.Error(err)
		os.Exit(1)
	}

	logging.Log.Info("Packster is up and running!")
}

func registerGitalbEndpoints(cfg config.Config, api *gin.RouterGroup, userRepo repository.IUserRepo) {
	if cfg.Gitlab == nil {
		return
	}

	handler := gitlab.NewGitlabHandler(cfg, userRepo)
	auth := api.Group("/auth/gitlab")
	{
		auth.GET("/redirect", handler.HandleRedirect)
		auth.GET("/callback", handler.HandleCallback)
	}

	api.GET("/auth/session", handler.HandleSession)
	api.GET("/user/candidates", handler.HandleListCandidates)
}

func registerProjectEndpoints(
	cfg config.Config,
	api *gin.RouterGroup,
	userRepo repository.IUserRepo,
	projectRepo repository.IProjectRepo,
	permRepo repository.IPermissionRepo,
	productRepo repository.IProductRepo,
	versionRepo repository.IVersionRepo,
) {
	handler := projects.NewProjectsHandler(cfg, userRepo, projectRepo, permRepo, productRepo, versionRepo)

	api.GET("/user/projects", handler.HandleListImported)
	api.POST("/user/projects", handler.HandleImport)
	api.DELETE("/projects/:id", handler.HandleDeleteProject)

	api.GET("/projects/:id/permissions", handler.HandleListPermissions)
	api.PUT("/projects/:id/permissions", handler.HandleSetPermission)
	api.DELETE("/projects/:id/permissions/:user_id", handler.HandleRevokePermission)
	api.GET("/projects/:id/permissions/candidates", handler.HandleSearchUsers)

	api.GET("/projects/:id/products", handler.HandleListProducts)
	api.POST("/projects/:id/products", handler.HandleCreateProduct)
	api.DELETE("/projects/:id/products/:product_id", handler.HandleDeleteProduct)

	api.GET("/products/:product_id/versions", handler.HandleListVersions)
	api.POST("/products/:product_id/versions", handler.HandleUploadVersion)
	api.GET("/versions/:version_id", handler.HandleDownloadVersion)
	api.DELETE("/versions/:version_id", handler.HandleDeleteVersion)

	api.GET("/projects/:id/products/:product_name/versions/:version_name", handler.HandleDownloadByName)
}

func printBanner() {
	fmt.Print(PURPLE)
	fmt.Print(BANNER)
	fmt.Print(RESET)
	fmt.Println()

	buildTime := BUILD_TIME
	if buildTime == "" {
		buildTime = "unknown"
	}

	fmt.Printf("\t\t%s • %s\n\n", MAINTAINER, buildTime)
}
