package service

import (
	"context"
	"strings"
	"time"

	redisrepo "mini_im/backend/internal/repository/redis"
)

type OnlineService struct {
	onlineRepo *redisrepo.OnlineRepository
	serverID   string
	ttl        time.Duration
}

func NewOnlineService(onlineRepo *redisrepo.OnlineRepository, serverID string, ttl time.Duration) *OnlineService {
	serverID = strings.TrimSpace(serverID)
	if serverID == "" {
		serverID = "ws-1"
	}
	if ttl <= 0 {
		ttl = 60 * time.Second
	}

	return &OnlineService{
		onlineRepo: onlineRepo,
		serverID:   serverID,
		ttl:        ttl,
	}
}

func (s *OnlineService) MarkOnline(ctx context.Context, userID int64, connectedAt time.Time) error {
	return s.onlineRepo.SetOnline(ctx, userID, s.serverID, connectedAt, s.ttl)
}

func (s *OnlineService) RefreshOnline(ctx context.Context, userID int64) error {
	if err := s.onlineRepo.RefreshOnlineTTL(ctx, userID, s.ttl); err != nil {
		if err == redisrepo.ErrOnlineStatusNotFound {
			return s.MarkOnline(ctx, userID, time.Now())
		}
		return err
	}
	return nil
}

func (s *OnlineService) MarkOffline(ctx context.Context, userID int64) error {
	return s.onlineRepo.DeleteOnline(ctx, userID)
}
