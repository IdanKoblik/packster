package config

import (
	"os"
	"strings"
	"fmt"

	"artifactor/internal/logging"
	"artifactor/pkg/config"
	"github.com/goccy/go-yaml"
)

func ParseConfig(path string) (config.Config, error) {
	logging.Log.Debugf("Conifg path: %s\n", path)

	var cfg config.Config
	if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
		return cfg, fmt.Errorf("Unsupported config file type")
	}

	data, err := os.ReadFile(path); if err != nil {
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
