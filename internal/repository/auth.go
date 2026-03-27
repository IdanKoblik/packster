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
	ListTokens() ([]string, error)
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

// tokenDoc is the document stored in MongoDB — only the hashed ID.
type tokenDoc struct {
	ID string `bson:"_id"`
}

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
	claims, err := utils.ParseToken(rawToken, r.Cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	hashedID := utils.Hash(claims.Subject)
	_, err = r.RedisClient.Get(context.Background(), r.getCacheKey(hashedID)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	if r.MockError != nil {
		return nil, r.MockError
	}

	r.RedisClient.Set(context.Background(), r.getCacheKey(hashedID), "_", r.CacheTTL)

	return r.MockToken, nil
}

func (r *AuthRepository) CreateToken(request *types.RegisterRequest) (string, error) {
	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.TokenCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := uuid.NewString()
	hashedID := utils.Hash(id)

	_, err := collection.InsertOne(ctx, tokenDoc{ID: hashedID})
	if err != nil {
		return "", err
	}

	r.RedisClient.Set(context.Background(), r.getCacheKey(hashedID), "_", r.CacheTTL)

	jwtToken, err := utils.SignToken(id, request.Admin, r.Cfg.JWTSecret)
	if err != nil {
		return "", err
	}

	return jwtToken, nil
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
	claims, err := utils.ParseToken(rawToken, r.Cfg.JWTSecret)
	if err != nil {
		return err
	}

	hashedID := utils.Hash(claims.Subject)
	r.RedisClient.Del(context.Background(), r.getCacheKey(hashedID))

	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.TokenCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = collection.DeleteOne(ctx, bson.M{"_id": hashedID})
	if err != nil {
		return err
	}

	return nil
}

func (r *AuthRepository) ListTokens() ([]string, error) {
	collection := r.MongoDatabase.Collection(r.Cfg.Mongo.TokenCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var docs []tokenDoc
	if err := cursor.All(context.Background(), &docs); err != nil {
		return nil, err
	}

	hashes := make([]string, len(docs))
	for i, doc := range docs {
		hashes[i] = doc.ID
	}

	return hashes, nil
}

func (r *AuthRepository) FetchToken(rawToken string) (*types.ApiToken, error) {
	claims, err := utils.ParseToken(rawToken, r.Cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	hashedID := utils.Hash(claims.Subject)
	_, err = r.RedisClient.Get(context.Background(), r.getCacheKey(hashedID)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	if err == nil {
		metrics.AuthCacheHits.Inc()
	} else {
		metrics.AuthCacheMisses.Inc()

		collection := r.MongoDatabase.Collection(r.Cfg.Mongo.TokenCollection)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var doc tokenDoc
		err = collection.FindOne(ctx, bson.M{"_id": hashedID}).Decode(&doc)
		if err != nil {
			return nil, err
		}

		r.RedisClient.Set(context.Background(), r.getCacheKey(hashedID), "_", r.CacheTTL)
	}

	return &types.ApiToken{
		Token: hashedID,
		Admin: claims.Admin,
	}, nil
}
