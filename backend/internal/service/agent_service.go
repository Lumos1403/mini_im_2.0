package service

import "context"

type AgentService struct{}

func NewAgentService() *AgentService {
	return &AgentService{}
}

func (s *AgentService) EnsureDefaultAgentFriend(ctx context.Context, userID int64) error {
	return nil
}
