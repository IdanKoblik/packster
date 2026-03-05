package redis

import (
	"context"
	"fmt"
	"artifactor/pkg/config"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func OpenConnection(cfg *config.RedisConfig) error {
	Client = nil
	if cfg == nil {
		return fmt.Errorf("Missing redis config")
	}

	Client = redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
		Password: cfg.Password,
		DB: cfg.DB,
	})

	if err := Client.Ping(context.Background()).Err(); err != nil {
		Client = nil
		return fmt.Errorf("Failed to ping redis: %v", err)
	}

	return nil
}
