package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"artifactor/internal/utils"
	"artifactor/pkg/http"
	"artifactor/pkg/tokens"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type AuthRepoInterface interface {
	TokenExists(rawToken string) (bool, error)
	CreateToken(request *http.CreateRequest) (string, error)
	PruneToken(rawToken string) error
	IsAdmin(rawToken string) (bool, error)
	FetchToken(rawToken string) (*tokens.Token, error)
}

type AuthRepository struct {
	Rdb       *redis.Client
	SqlClient *pgx.Conn
	CacheTTL  time.Duration
}

const (
	REDIS_TOKEN_PREFIX = "token:"
)

var (
	ctx = context.Background()
)

func NewAuthRepository(redisClient *redis.Client, sqlClient *pgx.Conn) *AuthRepository {
	return &AuthRepository{
		Rdb:       redisClient,
		SqlClient: sqlClient,
		CacheTTL:  5 * time.Minute,
	}
}

func (r *AuthRepository) getCacheKey(token string) string {
	return REDIS_TOKEN_PREFIX + token
}

func (r *AuthRepository) CreateToken(request *http.CreateRequest) (string, error) {
	query := `INSERT INTO "users"("token", "admin", "upload", "delete")
				VALUES ($1, $2, $3, $4)`

	token := uuid.NewString()
	_, err := r.SqlClient.Exec(
		ctx,
		query,
		utils.Hash(token),
		request.Admin,
		request.Upload,
		request.Delete,
	)

	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	key := r.getCacheKey(utils.Hash(token))
	err = r.Rdb.Set(ctx, key, "true", r.CacheTTL).Err()
	if err != nil {
		return "", err
	}

	return token, nil
}

func (r *AuthRepository) IsAdmin(rawToken string) (bool, error) {
	exists, err := r.TokenExists(rawToken)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, fmt.Errorf("Api token not found")
	}

	admin := false
	err = r.SqlClient.QueryRow(
		ctx,
		`SELECT admin FROM users WHERE token = $1`,
		utils.Hash(rawToken),
	).Scan(&admin)

	if err != nil {
		return false, err
	}

	return admin, nil
}

func (r *AuthRepository) TokenExists(rawToken string) (bool, error) {
	key := r.getCacheKey(utils.Hash(rawToken))
	_, err := r.Rdb.Get(ctx, key).Result()
	if err == nil {
		return true, nil
	}

	query := `SELECT 1 FROM users WHERE token = $1`
	var exists int

	err = r.SqlClient.QueryRow(ctx, query, utils.Hash(rawToken)).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	err = r.Rdb.Set(ctx, key, "true", r.CacheTTL).Err()
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *AuthRepository) PruneToken(rawToken string) error {
	exists, err := r.TokenExists(rawToken)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("%s does not exists", rawToken)
	}

	key := r.getCacheKey(utils.Hash(rawToken))
	_, err = r.Rdb.Del(ctx, key).Result()
	if err != nil {
		return err
	}

	query := `DELETE FROM "users" WHERE "token" = $1`
	_, err = r.SqlClient.Exec(ctx, query, utils.Hash(rawToken))
	if err != nil {
		return err
	}

	return nil
}

func (r *AuthRepository) FetchToken(rawToken string) (*tokens.Token, error) {
	exists, err := r.TokenExists(rawToken)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("%s api token does not exists", rawToken)
	}

	query := `SELECT token, admin, upload, delete FROM users WHERE token = $1`

	var token tokens.Token
	err = r.SqlClient.QueryRow(ctx, query, utils.Hash(rawToken)).Scan(&token.Data, &token.Permissions.Admin, &token.Permissions.Upload, &token.Permissions.Delete)
	if err != nil {
		return nil, err
	}

	return &token, nil
}
