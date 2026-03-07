package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"

	"artifactor/internal/config"
	"artifactor/internal/endpoints"
	"artifactor/internal/logging"
	"artifactor/internal/middleware"
	"artifactor/internal/redis"
	"artifactor/internal/repository"
	"artifactor/internal/sql"

	"github.com/gin-gonic/gin"
)

const BANNER = `
 █████╗ ██████╗ ████████╗██╗███████╗ █████╗  ██████╗████████╗ ██████╗ ██████╗
██╔══██╗██╔══██╗╚══██╔══╝██║██╔════╝██╔══██╗██╔════╝╚══██╔══╝██╔═══██╗██╔══██╗
███████║██████╔╝   ██║   ██║█████╗  ███████║██║        ██║   ██║   ██║██████╔╝
██╔══██║██╔══██╗   ██║   ██║██╔══╝  ██╔══██║██║        ██║   ██║   ██║██╔══██╗
██║  ██║██║  ██║   ██║   ██║██║     ██║  ██║╚██████╗   ██║   ╚██████╔╝██║  ██║
╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ╚═╝╚═╝     ╚═╝  ╚═╝ ╚═════╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝
`

const MAINTAINER = "Idan Koblik"

const PURPLE = "\033[38;2;87;87;232m"
const RESET = "\033[0m"

var BUILD_TIME string

func main() {
	logging.SetupLogger()
	printBanner()

	cfg, err := config.ParseConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		logging.Log.Error(err)
		os.Exit(1)
	}

	logging.Log.Debugf("Max file size that can be uploaded: %d MB\n", cfg.FileUploadLimit)
	logging.Log.Info("Connecting to pgsql database.")
	logging.Log.Debugf("Username: %s", cfg.Sql.Username)
	logging.Log.Debugf("Password: %s", generatePasswordMask())
	logging.Log.Debugf("Addr: %s", cfg.Sql.Addr)
	logging.Log.Debugf("Database: %s\n", cfg.Sql.Database)

	err = sql.OpenConnection(&cfg.Sql)
	if err != nil {
		logging.Log.Error("Failed to connect to pgsql database\n", err)
		os.Exit(1)
	}

	defer sql.Conn.Close(context.Background())
	logging.Log.Info("Successfully connected to pgsql database!\n")

	logging.Log.Info("Connecting to redis database.")
	logging.Log.Debugf("Addr: %s", cfg.Redis.Addr)
	logging.Log.Debugf("Password: %s", generatePasswordMask())

	err = redis.OpenConnection(&cfg.Redis)
	if err != nil {
		logging.Log.Error("Failed to connect to redis database\n", err)
		os.Exit(1)
	}

	defer redis.Client.Close()
	logging.Log.Info("Successfully connected to redis database!\n")

	logging.Log.Info("Starting rest api")
	router := gin.Default()

	authRepo := repository.NewAuthRepository(redis.Client, sql.Conn)
	authHandler := endpoints.NewAuthHandler(authRepo)

	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(authRepo))
	{
		api.PUT("/register", authHandler.HandleRegister)
		api.DELETE("/prune/:token", authHandler.HandlePrune)
		api.GET("/fetch/:token", authHandler.HandleFetch)
	}

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = "0.0.0.0:8080"
	}

	if err := router.Run(addr); err != nil {
		logging.Log.Error("Failed to start rest api\n", err)
		os.Exit(1)
	}
}

func generatePasswordMask() string {
	n := rand.N(18) + 5
	return strings.Repeat("*", n)
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
