package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var ErrRecallEditCacheNotFound = errors.New("recall edit cache not found")

type MessageRepository struct {
	client *goredis.Client
}

func NewMessageRepository(client *goredis.Client) *MessageRepository {
	return &MessageRepository{client: client}
}

func (r *MessageRepository) SaveRecallEditCache(ctx context.Context, messageID int64, userID int64, content string, ttl time.Duration) error {
	if ttl <= 0 {
		return errors.New("recall edit cache ttl must be positive")
	}
	return r.client.Set(ctx, recallEditKey(messageID, userID), content, ttl).Err()
}

func (r *MessageRepository) GetRecallEditCache(ctx context.Context, messageID int64, userID int64) (string, error) {
	value, err := r.client.Get(ctx, recallEditKey(messageID, userID)).Result()
	if errors.Is(err, goredis.Nil) {
		return "", ErrRecallEditCacheNotFound
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *MessageRepository) DeleteRecallEditCache(ctx context.Context, messageID int64, userID int64) error {
	return r.client.Del(ctx, recallEditKey(messageID, userID)).Err()
}

func recallEditKey(messageID int64, userID int64) string {
	return fmt.Sprintf("im:message:recall_edit:%s:%s", strconv.FormatInt(messageID, 10), strconv.FormatInt(userID, 10))
}
