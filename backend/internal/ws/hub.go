package ws

import (
	"context"
	"errors"
	"time"

	"mini_im/backend/internal/pkg/logger"

	"go.uber.org/zap"
)

var ErrSendQueueFull = errors.New("websocket send queue full")
var ErrClientClosed = errors.New("websocket client closed")

type OnlineStatusService interface {
	MarkOnline(ctx context.Context, userID int64, connectedAt time.Time) error
	RefreshOnline(ctx context.Context, userID int64) error
	MarkOffline(ctx context.Context, userID int64) error
}

type Hub struct {
	clients    map[int64]map[*Client]struct{}
	register   chan registerRequest
	unregister chan unregisterRequest
	sendToUser chan sendToUserRequest
	online     OnlineStatusService
}

type registerRequest struct {
	ctx    context.Context
	client *Client
	done   chan error
}

type unregisterRequest struct {
	ctx    context.Context
	client *Client
	done   chan error
}

type sendToUserRequest struct {
	ctx     context.Context
	userID  int64
	payload []byte
	done    chan error
}

func NewHub(online OnlineStatusService) *Hub {
	return &Hub{
		clients:    make(map[int64]map[*Client]struct{}),
		register:   make(chan registerRequest),
		unregister: make(chan unregisterRequest),
		sendToUser: make(chan sendToUserRequest),
		online:     online,
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.closeAllClients()
			return
		case req := <-h.register:
			req.done <- h.handleRegister(req.ctx, req.client)
		case req := <-h.unregister:
			req.done <- h.handleUnregister(req.ctx, req.client)
		case req := <-h.sendToUser:
			req.done <- h.handleSendToUser(req.ctx, req.userID, req.payload)
		}
	}
}

func (h *Hub) Register(ctx context.Context, client *Client) error {
	done := make(chan error, 1)
	select {
	case h.register <- registerRequest{ctx: ctx, client: client, done: done}:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *Hub) Unregister(ctx context.Context, client *Client) error {
	done := make(chan error, 1)
	select {
	case h.unregister <- unregisterRequest{ctx: ctx, client: client, done: done}:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *Hub) SendToUser(ctx context.Context, userID int64, payload []byte) error {
	done := make(chan error, 1)
	select {
	case h.sendToUser <- sendToUserRequest{ctx: ctx, userID: userID, payload: payload, done: done}:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *Hub) RefreshOnline(ctx context.Context, userID int64) error {
	if h.online == nil {
		return nil
	}
	return h.online.RefreshOnline(ctx, userID)
}

func (h *Hub) handleRegister(ctx context.Context, client *Client) error {
	if client == nil || client.UserID <= 0 {
		return errors.New("invalid websocket client")
	}
	if h.online != nil {
		if err := h.online.MarkOnline(ctx, client.UserID, time.Now()); err != nil {
			return err
		}
	}

	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[*Client]struct{})
	}
	h.clients[client.UserID][client] = struct{}{}
	logger.L().Info("websocket client registered", zap.Int64("user_id", client.UserID))
	return nil
}

func (h *Hub) handleUnregister(ctx context.Context, client *Client) error {
	if client == nil {
		return nil
	}

	userClients, exists := h.clients[client.UserID]
	if !exists {
		return nil
	}
	if _, exists := userClients[client]; !exists {
		return nil
	}

	delete(userClients, client)
	client.closeSend()
	if len(userClients) == 0 {
		delete(h.clients, client.UserID)
		if h.online != nil {
			if err := h.online.MarkOffline(ctx, client.UserID); err != nil {
				logger.L().Warn("websocket offline cleanup failed", zap.Int64("user_id", client.UserID), zap.Error(err))
				return err
			}
		}
	}

	logger.L().Info("websocket client unregistered", zap.Int64("user_id", client.UserID))
	return nil
}

func (h *Hub) handleSendToUser(ctx context.Context, userID int64, payload []byte) error {
	userClients := h.clients[userID]
	for client := range userClients {
		if err := client.enqueue(payload); err != nil {
			logger.L().Warn("websocket send queue unavailable", zap.Int64("user_id", userID), zap.Error(err))
			if err := h.handleUnregister(ctx, client); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *Hub) closeAllClients() {
	for userID, userClients := range h.clients {
		for client := range userClients {
			client.closeSend()
		}
		delete(h.clients, userID)
	}
}
