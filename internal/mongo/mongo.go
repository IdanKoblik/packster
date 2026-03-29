package mongo

import (
	"packster/pkg/config"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func OpenConnection(cfg *config.MongoConfig) (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(cfg.ConnectionString))
	if err != nil {
		return nil, err
	}

	err = CheckHealth(client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func CheckHealth(client *mongo.Client) error {
	if client == nil {
		return errors.New("Missing mongo client")
	}

	err := client.Ping(context.Background(), nil)
	if err != nil {
		return err
	}

	return nil
}
