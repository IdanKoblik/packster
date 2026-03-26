package repository

import (
	"artifactor/internal/metrics"
	"artifactor/internal/utils"
	"artifactor/pkg/types"
	"context"
	"errors"
	"time"

	"artifactor/pkg/config"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type IAuthRepo interface {
	TokenExists(rawToken string) (bool, error)
	CreateToken(request *types.RegisterRequest) (string, error)
	PruneToken(rawToken string) error
	IsAdmin(rawToken string) (bool, error)
	FetchToken(rawToken string) (*types.ApiToken, error)
	ListTokens() ([]types.ApiToken, error)
}

type AuthRepository struct {
	RedisClient   *redis.Client
	MongoClient   *mongo.Client
	MongoDatabase *mongo.Database
	Cfg           *config.Config
	CacheTTL      time.Duration
}

const (
	REDIS_TOKEN_PREFIX = "token:"
)

func NewAuthRepository(redisClient *redis.Client, mongoClient *mongo.Client, cfg *config.Config) *AuthRepository {
	return &AuthRepository{
		RedisClient:   redisClient,
		MongoClient:   mongoClient,
		MongoDatabase: mongoClient.Database(cfg.Mongo.Database),
		Cfg:           cfg,
		CacheTTL:      5 * time.Minute,
	}
}

func (r *AuthRepository) getCacheKey(token string) string {
	return REDIS_TOKEN_PREFIX + token
}

type TestableAuthRepository struct {
	*AuthRepository
	MockToken *types.ApiToken
	MockError error
}

func (r *TestableAuthRepository) SetMockToken(token *types.ApiToken) {
	r.MockToken = token
}

func (r *TestableAuthRepository) SetMockError(err error) {
	r.MockError = err
}

func (r *TestableAuthRepository) FetchToken(rawToken string) (*types.ApiToken, error) {
	hashedToken := utils.Hash(rawToken)
	_, err := r.RedisClient.Get(context.Background(), r.getCacheKey(hashedToken)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	if r.MockError != nil {
		return nil, r.MockError
	}

	r.RedisClient.Set(context.Background(), r.getCacheKey(hashedToken), "_", r.CacheTTL)

	return r.MockToken, nil
}

func (r *AuthRepository) CreateToken(request *types.RegisterRequest) (string, error) {
	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.TokenCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token := uuid.NewString()
	hashedToken := utils.Hash(token)
	apiToken := types.ApiToken{
		Token: hashedToken,
		Admin: request.Admin,
	}

	_, err := collection.InsertOne(ctx, apiToken)
	if err != nil {
		return "", err
	}

	r.RedisClient.Set(context.Background(), r.getCacheKey(hashedToken), "_", r.CacheTTL)

	return token, nil
}

func (r *AuthRepository) IsAdmin(rawToken string) (bool, error) {
	token, err := r.FetchToken(rawToken)
	if err != nil {
		return false, err
	}

	return token.Admin, nil
}

func (r *AuthRepository) TokenExists(rawToken string) (bool, error) {
	token, err := r.FetchToken(rawToken)
	if err != nil {
		return false, err
	}

	return token != nil, nil
}

func (r *AuthRepository) PruneToken(rawToken string) error {
	hashedToken := utils.Hash(rawToken)
	_, err := r.RedisClient.Get(context.Background(), r.getCacheKey(hashedToken)).Result()
	if err != nil {
		return err
	}

	r.RedisClient.Del(context.Background(), r.getCacheKey(hashedToken))
	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.TokenCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = collection.DeleteOne(ctx, bson.M{"_id": utils.Hash(rawToken)})
	if err != nil {
		return err
	}

	return nil
}

func (r *AuthRepository) ListTokens() ([]types.ApiToken, error) {
	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.TokenCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var tokens []types.ApiToken
	if err := cursor.All(context.Background(), &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *AuthRepository) FetchToken(rawToken string) (*types.ApiToken, error) {
	hashedToken := utils.Hash(rawToken)
	_, err := r.RedisClient.Get(context.Background(), r.getCacheKey(hashedToken)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	if err == nil {
		metrics.AuthCacheHits.Inc()
	} else {
		metrics.AuthCacheMisses.Inc()
	}

	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.TokenCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var apiToken types.ApiToken
	err = collection.FindOne(ctx, bson.M{"_id": utils.Hash(rawToken)}).Decode(&apiToken)
	if err != nil {
		return nil, err
	}

	r.RedisClient.Set(context.Background(), r.getCacheKey(hashedToken), "_", r.CacheTTL)

	return &apiToken, nil
}
