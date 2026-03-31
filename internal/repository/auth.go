package repository

import (
	"context"
	"database/sql"
	"errors"
	"packster/internal/metrics"
	"packster/internal/utils"
	"packster/pkg/config"
	"packster/pkg/types"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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
	RedisClient *redis.Client
	DB          *sql.DB
	Cfg         *config.Config
	CacheTTL    time.Duration
}

const (
	REDIS_TOKEN_PREFIX = "token:"
)

func NewAuthRepository(redisClient *redis.Client, db *sql.DB, cfg *config.Config) *AuthRepository {
	return &AuthRepository{
		RedisClient: redisClient,
		DB:          db,
		Cfg:         cfg,
		CacheTTL:    5 * time.Minute,
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
	token := uuid.NewString()
	hashedToken := utils.Hash(token)

	ctx, cancel := dbCtx()
	defer cancel()

	result, err := r.DB.ExecContext(ctx,
		"INSERT INTO principals (type, admin) VALUES ('token', ?)",
		request.Admin)
	if err != nil {
		return "", err
	}

	principalID, err := result.LastInsertId()
	if err != nil {
		return "", err
	}

	_, err = r.DB.ExecContext(ctx,
		"INSERT INTO api_tokens (id, token_hash) VALUES (?, ?)",
		principalID, hashedToken)
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

	ctx, cancel := dbCtx()
	defer cancel()

	_, err = r.DB.ExecContext(ctx,
		`DELETE p FROM principals p
		 JOIN api_tokens at ON p.id = at.id
		 WHERE at.token_hash = ?`,
		hashedToken)
	return err
}

func (r *AuthRepository) ListTokens() ([]types.ApiToken, error) {
	ctx, cancel := dbCtx()
	defer cancel()

	rows, err := r.DB.QueryContext(ctx,
		"SELECT at.token_hash, p.admin FROM api_tokens at JOIN principals p ON at.id = p.id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []types.ApiToken
	for rows.Next() {
		var t types.ApiToken
		if err := rows.Scan(&t.Token, &t.Admin); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
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

	ctx, cancel := dbCtx()
	defer cancel()

	var apiToken types.ApiToken
	err = r.DB.QueryRowContext(ctx,
		`SELECT at.token_hash, p.admin
		 FROM api_tokens at
		 JOIN principals p ON at.id = p.id
		 WHERE at.token_hash = ?`,
		hashedToken).Scan(&apiToken.Token, &apiToken.Admin)
	if err != nil {
		return nil, err
	}

	r.RedisClient.Set(context.Background(), r.getCacheKey(hashedToken), "_", r.CacheTTL)

	return &apiToken, nil
}
