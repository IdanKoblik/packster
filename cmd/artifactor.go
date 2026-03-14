package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"

	internalconfig "artifactor/internal/config"
	"artifactor/internal/endpoints"
	"artifactor/internal/endpoints/auth"
	"artifactor/internal/endpoints/product"
	"artifactor/internal/flags"
	"artifactor/internal/logging"
	"artifactor/internal/middleware"
	internalmongo "artifactor/internal/mongo"
	internalredis "artifactor/internal/redis"
	"artifactor/internal/repository"
	"artifactor/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
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

	cfg, err := internalconfig.ParseConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		logging.Log.Error(err)
		os.Exit(1)
	}

	logging.Log.Debugf("Max file size that can be uploaded: %d MB\n", cfg.FileUploadLimit)
	logging.Log.Info("Connecting to mongo database.")
	logging.Log.Debugf("Connection URL: %s", generateMask())
	logging.Log.Debugf("Database: %s", cfg.Mongo.Database)

	mongoClient, err := internalmongo.OpenConnection(&cfg.Mongo)
	if err != nil {
		logging.Log.Error("Failed to connect to mongo database\n", err)
		os.Exit(1)
	}

	defer mongoClient.Disconnect(context.Background())
	logging.Log.Info("Successfully connected to mongo database!\n")

	logging.Log.Info("Connecting to redis database.")
	logging.Log.Debugf("Addr: %s", cfg.Redis.Addr)
	logging.Log.Debugf("Password: %s", generateMask())

	redisClient, err := internalredis.OpenConnection(&cfg.Redis)
	if err != nil {
		logging.Log.Error("Failed to connect to redis database\n", err)
		os.Exit(1)
	}

	defer redisClient.Close()
	logging.Log.Info("Successfully connected to redis database!\n")

	authRepo := repository.NewAuthRepository(redisClient, mongoClient, &cfg)

	logging.Log.Info("Starting rest api")
	router := gin.Default()

	api := router.Group("/api")

	setupAuthEndpoints(authRepo, redisClient, mongoClient, api)
	setupProductEndpoints(authRepo, mongoClient, &cfg, api)

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = "0.0.0.0:8080"
	}

	if err := router.Run(addr); err != nil {
		logging.Log.Error("Failed to start rest api\n", err)
		os.Exit(1)
	}
}

func setupAuthEndpoints(authRepo *repository.AuthRepository, redisClient *redis.Client, mongoClient *mongo.Client, api *gin.RouterGroup) {
	authHandler := auth.NewAuthHandler(authRepo)

	startFlagSystem(authRepo)
	if len(os.Args) > 1 {
		flag, err := flags.GetFlag(os.Args[1])
		if err == nil {
			err = flag.Handle(os.Args[1:])
			if err != nil {
				logging.Log.Error(err)
			}
		}
	}

	api.Use(middleware.AuthMiddleware(authRepo))
	{
		api.PUT("/register", authHandler.HandleRegister)
		api.DELETE("/prune/:token", authHandler.HandlePrune)
		api.GET("/fetch/:token", authHandler.HandleFetch)
		api.GET("/health", func(c *gin.Context) {
			endpoints.HandleHealth(c, mongoClient, redisClient)
		})
	}
}

func setupProductEndpoints(authRepo repository.IAuthRepo, mongoClient *mongo.Client, cfg *config.Config, api *gin.RouterGroup) {
	productRepo := repository.NewProductRepository(mongoClient, cfg)
	productHandler := product.NewProductHandler(productRepo)

	productApi := api.Group("/product")
	productApi.Use(middleware.AuthMiddleware(authRepo))
	{
		productApi.PUT("/create", productHandler.HandleCreate)
		productApi.DELETE("/delete/:product", productHandler.HandleDelete)
		productApi.GET("/fetch/:product", productHandler.HandleFetch)
		productApi.DELETE("/modify/:action", productHandler.HandleModify)
		productApi.PUT("/modify/:action", productHandler.HandleModify)
	}
}

func generateMask() string {
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

func startFlagSystem(r *repository.AuthRepository) {
	flags.InitFlagRegistry()

	flags.RegisterFlag(flags.InitToken(r))
}
