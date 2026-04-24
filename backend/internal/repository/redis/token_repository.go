package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var ErrTokenNotFound = errors.New("token not found")

type TokenRepository struct {
	client *goredis.Client
}

func NewTokenRepository(client *goredis.Client) *TokenRepository {
	return &TokenRepository{client: client}
}

func (r *TokenRepository) SaveRefreshToken(ctx context.Context, userID int64, deviceID string, jti string, ttl time.Duration) error {
	return r.client.Set(ctx, refreshKey(userID, deviceID), jti, ttl).Err()
}

func (r *TokenRepository) GetRefreshTokenJTI(ctx context.Context, userID int64, deviceID string) (string, error) {
	value, err := r.client.Get(ctx, refreshKey(userID, deviceID)).Result()
	if errors.Is(err, goredis.Nil) {
		return "", ErrTokenNotFound
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *TokenRepository) DeleteRefreshToken(ctx context.Context, userID int64, deviceID string) error {
	return r.client.Del(ctx, refreshKey(userID, deviceID)).Err()
}

func (r *TokenRepository) BlacklistAccessToken(ctx context.Context, jti string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return r.client.Set(ctx, blacklistKey(jti), "1", ttl).Err()
}

func (r *TokenRepository) IsAccessTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	count, err := r.client.Exists(ctx, blacklistKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func refreshKey(userID int64, deviceID string) string {
	return fmt.Sprintf("im:auth:refresh:%s:%s", strconv.FormatInt(userID, 10), deviceID)
}

func blacklistKey(jti string) string {
	return "im:auth:blacklist:" + jti
}
