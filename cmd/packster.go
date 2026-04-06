package main

import (
	"fmt"
	"strings"
	"os"
	"math/rand/v2"

	"packster/internal/sql"
	"packster/internal/config"
	"packster/internal/logging"
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

	cfg, err := config.ParseConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		logging.Log.Error(err)
		os.Exit(1)
	}

	logging.Log.Debugf("Max file size that can be uploaded: %d MB\n", cfg.FileUploadLimit)

	logging.Log.Info("Connecting to mysql db:")
	logging.Log.Infof("Host: %s", cfg.Sql.Password)
	logging.Log.Infof("Database: %s", cfg.Sql.DB)
	logging.Log.Infof("Username: %s", cfg.Sql.Username)
	logging.Log.Infof("Password: %s", generateMask())

	err = sql.ConnectToMysql(&cfg.Sql)
	if err != nil {
		logging.Log.Error(err)
		os.Exit(1)
	}

	defer sql.MysqlConn.Close()
	logging.Log.Info("Successfully connected to mysql db")

	if cfg.Gitlab != nil {
		logging.Log.Info("Gitlab sso detected")
		logging.Log.Infof("Host: %s", cfg.Gitlab.Host)
		logging.Log.Infof("Application ID: %s", generateMask())
		logging.Log.Infof("Secret: %s", generateMask())
	}

	logging.Log.Info("Packster is up and running!")
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
