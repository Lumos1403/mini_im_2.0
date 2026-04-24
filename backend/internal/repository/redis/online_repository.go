package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var ErrOnlineStatusNotFound = errors.New("online status not found")

type OnlineRepository struct {
	client *goredis.Client
}

type OnlineStatus struct {
	ServerID    string `json:"server_id"`
	ConnectedAt string `json:"connected_at"`
}

func NewOnlineRepository(client *goredis.Client) *OnlineRepository {
	return &OnlineRepository{client: client}
}

func (r *OnlineRepository) SetOnline(ctx context.Context, userID int64, serverID string, connectedAt time.Time, ttl time.Duration) error {
	value, err := json.Marshal(OnlineStatus{
		ServerID:    serverID,
		ConnectedAt: connectedAt.Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		return err
	}

	return r.client.Set(ctx, onlineKey(userID), value, ttl).Err()
}

func (r *OnlineRepository) RefreshOnlineTTL(ctx context.Context, userID int64, ttl time.Duration) error {
	refreshed, err := r.client.Expire(ctx, onlineKey(userID), ttl).Result()
	if err != nil {
		return err
	}
	if !refreshed {
		return ErrOnlineStatusNotFound
	}
	return nil
}

func (r *OnlineRepository) DeleteOnline(ctx context.Context, userID int64) error {
	return r.client.Del(ctx, onlineKey(userID)).Err()
}

func onlineKey(userID int64) string {
	return fmt.Sprintf("im:online:%s", strconv.FormatInt(userID, 10))
}
