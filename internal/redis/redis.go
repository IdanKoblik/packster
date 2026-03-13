package redis

import (
	"artifactor/pkg/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func OpenConnection(cfg *config.RedisConfig) (*redis.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Missing redis config")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	err := CheckHealth(client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func CheckHealth(client *redis.Client) error {
	if client == nil {
		return fmt.Errorf("Missing redis client")
	}

	err := client.Ping(context.Background()).Err()
	if err != nil {
		return fmt.Errorf("Failed to ping redis: %v", err)
	}

	return nil
}
