package config

import (
	"os"
	"strings"
	"fmt"

	"packster/internal/logging"
	"packster/pkg/config"
	"github.com/goccy/go-yaml"
)

func ParseConfig(path string) (config.Config, error) {
	var cfg config.Config
	if path == "" {
		return cfg, fmt.Errorf("Config file is required!")
	}

	logging.Log.Debugf("Conifg path: %s\n", path)

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
