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

	return token, nil
}

func (r *AuthRepository) FetchToken(rawToken string) (*tokens.Token, error) {
	key := r.getCacheKey(utils.Hash(rawToken))

	query := `SELECT token, admin, upload, delete FROM users WHERE token = $1`
	var token tokens.Token
	err := r.SqlClient.QueryRow(ctx, query, utils.Hash(rawToken)).Scan(&token.Data, &token.Permissions.Admin, &token.Permissions.Upload, &token.Permissions.Delete)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	err = r.Rdb.Set(ctx, key, "true", r.CacheTTL).Err()
	if err != nil {
		return nil, err
	}

	return &token, nil
}
