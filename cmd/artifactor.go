package main

import (
	"fmt"
	"os"
	"strings"

	"artifactor/internal/config"
	"artifactor/internal/logging"
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

	fmt.Print(PURPLE)
	fmt.Print(BANNER)
	fmt.Print(RESET)
	fmt.Println()

	buildTime := BUILD_TIME
	if buildTime == "" {
		buildTime = "unknown"
	}

	fmt.Printf("\t\t%s • %s\n\n", MAINTAINER, buildTime)

	cfg, err := config.ParseConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		logging.Log.Error(err)
		os.Exit(1)
	}

	logging.Log.Debugf("Max file size that can be uploaded: %d\n", cfg.FileUploadLimit)
	logging.Log.Info("Connecting to pgsql database.")
	logging.Log.Debugf("Username: %s", cfg.Sql.Username)
	logging.Log.Debugf("Password: %s", strings.Repeat("*", len(cfg.Sql.Password)))
	logging.Log.Debugf("Addr: %s", cfg.Sql.Addr)
	logging.Log.Debugf("Database: %s\n", cfg.Sql.Database)

	logging.Log.Info("Connecting to redis database.")
	logging.Log.Debugf("Addr: %s", cfg.Redis.Addr)
	logging.Log.Debugf("Password: %s", strings.Repeat("*", len(cfg.Redis.Password)))
}
