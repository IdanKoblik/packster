package logging

import (
	"os"

	"github.com/charmbracelet/log"
)
var Log *log.Logger = log.New(os.Stderr)

func SetupLogger() {
	Log = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller: true,
		ReportTimestamp: true,
		Prefix: "📦",
		Level: log.DebugLevel,
	})
}
