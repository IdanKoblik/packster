package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"artifactor/pkg/requests"
	"artifactor/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type AuthRepository struct {
	Rdb       *redis.Client
	SqlClient *pgx.Conn
	CacheTTL  time.Duration
}

const (
	REDIS_USER_PREFIX = "user"
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
	return REDIS_USER_PREFIX + token
}

func (r *AuthRepository) CreateUser(request *requests.RegisterRequest) error {
	query := `INSERT INTO "users"("username", "name", "mail", "password", "admin", "upload", "delete")
				VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.SqlClient.Exec(
		ctx,
		query,
		request.Username,
		request.Name,
		request.Mail,
		utils.Hash(request.Password),
		request.Admin,
		request.Permissions.Upload,
		request.Permissions.Delete,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *AuthRepository) UserExists(username string) (bool, error) {
	key := r.getCacheKey(username)
	val, err := r.Rdb.Get(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, err
	}

	exists := false
	if err == nil && val != "" {
		exists, err = strconv.ParseBool(val)
		if err != nil {
			return false, err
		}
	}

	if !exists {
		query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
		err = r.SqlClient.QueryRow(ctx, query, username).Scan(&exists)
		if err != nil {
			return false, err
		}
	}

	err = r.Rdb.Set(ctx, key, strconv.FormatBool(exists), r.CacheTTL).Err()
	if err != nil {
		return false, err
	}

	return exists, nil
}
