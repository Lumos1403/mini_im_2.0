package ws

import (
	"context"
	"strconv"
	"time"

	"mini_im/backend/internal/service"
)

type AgentNotifier struct {
	hub     *Hub
	timeout time.Duration
}

func NewAgentNotifier(hub *Hub, timeout time.Duration) *AgentNotifier {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &AgentNotifier{
		hub:     hub,
		timeout: timeout,
	}
}

func (n *AgentNotifier) NotifyAgentMessage(ctx context.Context, userID int64, data service.MessageReceiveOutput) error {
	return n.send(ctx, userID, "server-"+data.MessageID, EventChatMessageReceive, data)
}

func (n *AgentNotifier) NotifyAgentMessageStreamStart(ctx context.Context, userID int64, data service.AgentMessageStartOutput) error {
	return n.send(ctx, userID, "server-"+data.StreamID+"-start", EventAgentMessageStart, data)
}

func (n *AgentNotifier) NotifyAgentMessageStreamChunk(ctx context.Context, userID int64, data service.AgentMessageChunkOutput) error {
	return n.send(ctx, userID, "server-"+data.StreamID+"-chunk-"+strconv.Itoa(data.ChunkIndex), EventAgentMessageChunk, data)
}

func (n *AgentNotifier) NotifyAgentMessageStreamDone(ctx context.Context, userID int64, data service.AgentMessageDoneOutput) error {
	return n.send(ctx, userID, "server-"+data.Message.MessageID+"-agent-done", EventAgentMessageDone, data)
}

func (n *AgentNotifier) NotifyAgentMessageStreamError(ctx context.Context, userID int64, data service.AgentMessageErrorOutput) error {
	return n.send(ctx, userID, "server-"+data.StreamID+"-error", EventAgentMessageError, data)
}

func (n *AgentNotifier) send(ctx context.Context, userID int64, seq string, eventType string, data any) error {
	if n == nil || n.hub == nil || userID <= 0 {
		return nil
	}

	payload, err := MarshalEnvelope(seq, eventType, data)
	if err != nil {
		return err
	}

	notifyCtx, cancel := context.WithTimeout(ctx, n.timeout)
	defer cancel()
	return n.hub.SendToUser(notifyCtx, userID, payload)
}
