package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"

	internalconfig "artifactor/internal/config"
	"artifactor/internal/endpoints"
	"artifactor/internal/endpoints/auth"
	"artifactor/internal/endpoints/product"
	"artifactor/internal/flags"
	"artifactor/internal/logging"
	"artifactor/internal/metrics"
	"artifactor/internal/middleware"
	internalmongo "artifactor/internal/mongo"
	internalredis "artifactor/internal/redis"
	"artifactor/internal/repository"
	"artifactor/internal/ui"
	"artifactor/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

// @title           Artifactor API
// @version         1.0.0
// @description     Package version management service — store, retrieve, and manage versioned build artifacts.
// @BasePath        /api
// @securityDefinitions.apikey  ApiKeyAuth
// @in              header
// @name            X-Api-Token
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

	startHealthProbes(mongoClient, redisClient)

	logging.Log.Info("Starting rest api")
	router := gin.Default()
	router.Use(middleware.PrometheusMiddleware())
	router.MaxMultipartMemory = int64(cfg.FileUploadLimit) << 20

	api := router.Group("/api")

	setupAuthEndpoints(authRepo, redisClient, mongoClient, api)
	setupProductEndpoints(authRepo, mongoClient, &cfg, api)

	if isUIEnabled() {
		ui.SetupUI(authRepo, router)
	}

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = "0.0.0.0:8080"
	}

	startMetricsServer(&cfg)

	if err := router.Run(addr); err != nil {
		logging.Log.Error("Failed to start rest api\n", err)
		os.Exit(1)
	}
}

// startMetricsServer serves /metrics on a dedicated port (default :9091) so it
// can be firewalled independently from the main API.
func startMetricsServer(cfg *config.Config) {
	metricsAddr := cfg.Metrics.Addr
	if metricsAddr == "" {
		metricsAddr = "0.0.0.0:9091"
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	go func() {
		logging.Log.Infof("Metrics server listening on %s", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, mux); err != nil {
			logging.Log.Errorf("Metrics server error: %v", err)
		}
	}()
}

// startHealthProbes runs a background goroutine that updates the mongo_up and
// redis_up gauges every 15 seconds so dashboards reflect current dependency health.
func startHealthProbes(mongoClient *mongo.Client, redisClient *redis.Client) {
	probe := func() {
		if err := internalmongo.CheckHealth(mongoClient); err != nil {
			metrics.MongoUp.Set(0)
		} else {
			metrics.MongoUp.Set(1)
		}
		if err := internalredis.CheckHealth(redisClient); err != nil {
			metrics.RedisUp.Set(0)
		} else {
			metrics.RedisUp.Set(1)
		}
	}

	probe()
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			probe()
		}
	}()
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
		api.GET("/tokens", authHandler.HandleListTokens)
		api.GET("/health", func(c *gin.Context) {
			endpoints.HandleHealth(c, mongoClient, redisClient)
		})
	}
}

func setupProductEndpoints(authRepo repository.IAuthRepo, mongoClient *mongo.Client, cfg *config.Config, api *gin.RouterGroup) {
	productRepo := repository.NewProductRepository(mongoClient, cfg)
	productHandler := product.NewProductHandler(productRepo, cfg.FileUploadLimit)

	productApi := api.Group("/product")
	productApi.Use(middleware.AuthMiddleware(authRepo))
	{
		productApi.PUT("/create", productHandler.HandleCreate)
		productApi.DELETE("/delete/:product", productHandler.HandleDelete)
		productApi.GET("/fetch/:product", productHandler.HandleFetch)
		productApi.GET("/list", productHandler.HandleListProducts)
		productApi.GET("/access", productHandler.HandleAccess)
		productApi.DELETE("/modify/:action", productHandler.HandleModify)
		productApi.PUT("/modify/:action", productHandler.HandleModify)
		productApi.POST("/upload", productHandler.HandleUpload)
		productApi.GET("/download/:product/:version", productHandler.HandleDownload)
		productApi.DELETE("/delete/:product/:version", productHandler.HandleDeleteVersion)
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
	flags.RegisterFlag(flags.UIFlag())
}

func isUIEnabled() bool {
	for _, arg := range os.Args[1:] {
		if arg == "--ui" {
			return true
		}
	}
	return false
}
