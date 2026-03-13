package flags

import (
	"fmt"

	"artifactor/internal/logging"
	"artifactor/pkg/flags"
)

var Flags map[string]flags.Flag

func InitFlagRegistry() {
	Flags = make(map[string]flags.Flag)
	logging.Log.Info("Initialized flag registry")
}

func RegisterFlag(flag flags.Flag) {
	logging.Log.Infof("Registering %s flag. %s", flag.Name, flag.Usage())
	for _, v := range flag.Description {
		logging.Log.Info(v)
	}

	Flags[flag.Cmd] = flag
}

func GetFlag(cmd string) (*flags.Flag, error) {
	flag, ok := Flags[cmd]
	if !ok {
		return nil, fmt.Errorf("%s flag was not found", cmd)
	}

	return &flag, nil
}
