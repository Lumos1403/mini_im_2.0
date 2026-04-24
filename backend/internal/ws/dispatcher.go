package ws

import (
	"context"
	"encoding/json"
	"time"

	"mini_im/backend/internal/model"
	"mini_im/backend/internal/pkg/logger"
	"mini_im/backend/internal/service"

	"go.uber.org/zap"
)

type MessageDispatcher interface {
	Dispatch(ctx context.Context, client *Client, envelope *Envelope) error
}

type SyncMessageDispatcher struct {
	messageService *service.MessageService
	hub            *Hub
	timeout        time.Duration
}

type chatMessageSendData struct {
	ConversationID string          `json:"conversation_id"`
	ClientMsgID    string          `json:"client_msg_id"`
	MessageType    string          `json:"message_type"`
	Content        string          `json:"content"`
	ExtraJSON      json.RawMessage `json:"extra_json"`
}

func NewSyncMessageDispatcher(messageService *service.MessageService, hub *Hub, timeout time.Duration) *SyncMessageDispatcher {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &SyncMessageDispatcher{
		messageService: messageService,
		hub:            hub,
		timeout:        timeout,
	}
}

func (d *SyncMessageDispatcher) Dispatch(ctx context.Context, client *Client, envelope *Envelope) error {
	if d == nil || d.messageService == nil || d.hub == nil {
		return client.sendErrorEnvelope(envelope.Seq, "dispatcher_unavailable", "message dispatcher unavailable", envelope.Type)
	}

	if envelope.Type != EventChatMessageSend {
		return client.sendErrorEnvelope(envelope.Seq, "unsupported_event", "unsupported websocket event", envelope.Type)
	}

	var data chatMessageSendData
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		return client.sendEnvelope(envelope.Seq, EventChatMessageFailed, service.SendMessageFailedOutput{
			ConversationID: data.ConversationID,
			SendStatus:     model.MessageSendStatusFailed,
			Code:           "invalid_request",
			Message:        "invalid message payload",
		})
	}

	dispatchCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	result, appErr := d.messageService.SendTextMessage(dispatchCtx, service.SendTextMessageInput{
		SenderID:       client.UserID,
		ConversationID: data.ConversationID,
		ClientMsgID:    data.ClientMsgID,
		MessageType:    data.MessageType,
		Content:        data.Content,
		ExtraJSON:      data.ExtraJSON,
	})
	if appErr != nil {
		return client.sendEnvelope(envelope.Seq, EventChatMessageFailed, service.SendMessageFailedOutput{
			ClientMsgID:    data.ClientMsgID,
			ConversationID: data.ConversationID,
			SendStatus:     model.MessageSendStatusFailed,
			Code:           "internal_error",
			Message:        appErr.Message,
		})
	}
	if result == nil {
		return client.sendEnvelope(envelope.Seq, EventChatMessageFailed, service.SendMessageFailedOutput{
			ClientMsgID:    data.ClientMsgID,
			ConversationID: data.ConversationID,
			SendStatus:     model.MessageSendStatusFailed,
			Code:           "internal_error",
			Message:        "message send failed",
		})
	}
	if result.Failed != nil {
		return client.sendEnvelope(envelope.Seq, EventChatMessageFailed, result.Failed)
	}

	if result.Receive != nil && result.ReceiverID > 0 && !result.Duplicated {
		if err := d.sendReceive(dispatchCtx, result.ReceiverID, result.Receive); err != nil {
			logger.L().Warn("websocket message receive push failed",
				zap.Int64("sender_id", client.UserID),
				zap.Int64("receiver_id", result.ReceiverID),
				zap.Error(err),
			)
		}
	}
	return client.sendEnvelope(envelope.Seq, EventChatMessageAck, result.Ack)
}

func (d *SyncMessageDispatcher) sendReceive(ctx context.Context, receiverID int64, data *service.MessageReceiveOutput) error {
	payload, err := MarshalEnvelope("server-"+data.MessageID, EventChatMessageReceive, data)
	if err != nil {
		return err
	}
	return d.hub.SendToUser(ctx, receiverID, payload)
}
