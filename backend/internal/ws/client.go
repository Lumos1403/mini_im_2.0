package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"mini_im/backend/internal/pkg/logger"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub

	options Options
	once    sync.Once
}

func NewClient(userID int64, conn *websocket.Conn, hub *Hub, options Options) *Client {
	return &Client{
		UserID:  userID,
		Conn:    conn,
		Send:    make(chan []byte, options.SendBufferSize),
		Hub:     hub,
		options: options,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		unregisterCtx, cancel := context.WithTimeout(context.Background(), c.options.WriteWait)
		defer cancel()
		if err := c.Hub.Unregister(unregisterCtx, c); err != nil {
			logger.L().Warn("websocket unregister failed", zap.Int64("user_id", c.UserID), zap.Error(err))
		}
		_ = c.Conn.Close()
	}()

	c.Conn.SetReadLimit(c.options.MaxMessageBytes)
	_ = c.Conn.SetReadDeadline(time.Now().Add(c.options.PongWait))
	c.Conn.SetPongHandler(func(string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(c.options.PongWait)); err != nil {
			return err
		}
		refreshCtx, cancel := context.WithTimeout(context.Background(), c.options.WriteWait)
		defer cancel()
		if err := c.Hub.RefreshOnline(refreshCtx, c.UserID); err != nil {
			logger.L().Warn("websocket online ttl refresh failed", zap.Int64("user_id", c.UserID), zap.Error(err))
			return err
		}
		return nil
	})

	for {
		messageType, payload, err := c.Conn.ReadMessage()
		if err != nil {
			logger.L().Info("websocket read stopped", zap.Int64("user_id", c.UserID), zap.Error(err))
			return
		}
		if messageType != websocket.TextMessage {
			continue
		}

		var envelope Envelope
		if err := json.Unmarshal(payload, &envelope); err != nil {
			logger.L().Warn("websocket invalid envelope", zap.Int64("user_id", c.UserID), zap.Error(err))
			return
		}
		if !c.handleEnvelope(&envelope) {
			return
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(c.options.PingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case payload, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				logger.L().Info("websocket write stopped", zap.Int64("user_id", c.UserID), zap.Error(err))
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.L().Info("websocket ping failed", zap.Int64("user_id", c.UserID), zap.Error(err))
				return
			}
		}
	}
}

func (c *Client) handleEnvelope(envelope *Envelope) bool {
	switch envelope.Type {
	case EventPing:
		if err := c.sendEnvelope(envelope.Seq, EventPong, map[string]string{}); err != nil {
			logger.L().Warn("websocket pong enqueue failed", zap.Int64("user_id", c.UserID), zap.Error(err))
			return false
		}
	case EventPong:
		return true
	default:
		if err := c.sendErrorEnvelope(envelope.Seq, "unsupported_event", "unsupported websocket event", envelope.Type); err != nil {
			logger.L().Warn("websocket error enqueue failed", zap.Int64("user_id", c.UserID), zap.Error(err))
			return false
		}
	}
	return true
}

func (c *Client) sendEnvelope(seq string, eventType string, data any) error {
	payload, err := MarshalEnvelope(seq, eventType, data)
	if err != nil {
		return err
	}

	return c.enqueue(payload)
}

func (c *Client) sendErrorEnvelope(seq string, code string, message string, eventType string) error {
	return c.sendEnvelope(seq, EventError, ErrorData{
		Code:      code,
		Message:   message,
		EventType: eventType,
	})
}

func (c *Client) enqueue(payload []byte) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = ErrClientClosed
		}
	}()

	select {
	case c.Send <- payload:
		return nil
	default:
		return ErrSendQueueFull
	}
}

func (c *Client) closeSend() {
	c.once.Do(func() {
		close(c.Send)
	})
}
