package redis

import (
	"context"
	"errors"
	"time"

	"mini_im/backend/internal/config"

	goredis "github.com/redis/go-redis/v9"
)

func New(cfg config.RedisConfig) (*goredis.Client, error) {
	if cfg.Addr == "" {
		return nil, errors.New("REDIS_ADDR is required")
	}

	client := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}
