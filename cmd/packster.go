package main

import (
	"fmt"
	"strings"
	"os"
	"math/rand/v2"

	"packster/pkg/config"
	"packster/internal/sql"
	internalConfig "packster/internal/config"
	"packster/internal/logging"
	"packster/internal/endpoints"
	"packster/internal/endpoints/gitlab"
	"packster/internal/repository"
	"packster/internal/ui"

	"github.com/gin-gonic/gin"
)

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

	logging.Log.Info("Connecting to pgsql db:")
	logging.Log.Infof("Host: %s", cfg.Sql.Host)
	logging.Log.Infof("Port: %d", cfg.Sql.Port)
	logging.Log.Infof("Username: %s", cfg.Sql.User)
	logging.Log.Infof("Password: %s", generateMask())

	_, err = sql.OpenPgsqlConnection(&cfg.Sql)
	if err != nil {
		logging.Log.Error(err)
		os.Exit(1)
	}

	defer sql.PgsqlConn.Close()
	logging.Log.Info("Successfully connected to pgsql db")

	if cfg.Gitlab != nil {
		logging.Log.Info("Gitlab sso detected")
		logging.Log.Infof("Host: %s", cfg.Gitlab.Host)
		logging.Log.Infof("Application ID: %s", generateMask())
		logging.Log.Infof("Secret: %s", generateMask())
	}

	logging.Log.Info("Setting up rest api")
	router := gin.Default()

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "0.0.0.0:8080"
	}

	logging.Log.Debugf("Addr: %s", addr)

	repo := repository.NewAccountRepo(sql.PgsqlConn)
	api := router.Group("/api")
	{
		api.GET("/health", func(c *gin.Context){
			endpoints.HandleHealth(c, sql.PgsqlConn)
		})
	}

	auth := api.Group("/auth")

	if cfg.Gitlab != nil {
		logging.Log.Info("Setting up gitlab endpoints")
		setupGitlabEndpoints(auth, repo, &cfg)
	}

	ui.RegisterRoutes(router)

	router.Run(addr)
	logging.Log.Info("Packster is up and running!")
}

func setupGitlabEndpoints(authGroup *gin.RouterGroup, repo *repository.AccountRepo, cfg *config.Config) {
	handler := gitlab.NewGitlabHandler(repo, cfg)
	group := authGroup.Group("/gitlab")
	{
		group.GET("/status", handler.HandleStatus)
		group.GET("/redirect", handler.HandleRedirect)
		group.GET("/callback", handler.HandleCallback)
	}
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

func generateMask() string {
	n := rand.N(18) + 5
	return strings.Repeat("*", n)
}
