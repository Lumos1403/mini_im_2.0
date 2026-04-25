package ws

import (
	"context"
	"time"

	"mini_im/backend/internal/service"
)

type RecallNotifier struct {
	hub     *Hub
	timeout time.Duration
}

func NewRecallNotifier(hub *Hub, timeout time.Duration) *RecallNotifier {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &RecallNotifier{
		hub:     hub,
		timeout: timeout,
	}
}

func (n *RecallNotifier) NotifyMessageRecalled(ctx context.Context, recipientIDs []int64, data service.MessageRecalledEventOutput) error {
	if n == nil || n.hub == nil || len(recipientIDs) == 0 {
		return nil
	}

	payload, err := MarshalEnvelope("server-recall-"+data.MessageID, EventChatMessageRecalled, data)
	if err != nil {
		return err
	}

	notifyCtx, cancel := context.WithTimeout(ctx, n.timeout)
	defer cancel()

	var firstErr error
	for _, userID := range recipientIDs {
		if err := n.hub.SendToUser(notifyCtx, userID, payload); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
