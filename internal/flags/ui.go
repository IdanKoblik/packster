package flags

import (
	"packster/internal/logging"
	"packster/pkg/flags"
)

func UIFlag() flags.Flag {
	return flags.Flag{
		Cmd:  "--ui",
		Name: "ui",
		Args: []string{},
		Description: []string{
			"Enables the web UI served at /ui.",
		},
		Handle: func(args []string) error {
			logging.Log.Info("UI flag detected, web interface will be enabled")
			return nil
		},
	}
}
