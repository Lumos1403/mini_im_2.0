package ws

import (
	"encoding/json"
	"time"
)

const (
	EventPing  = "ping"
	EventPong  = "pong"
	EventError = "error"

	EventChatMessageSend     = "chat.message.send"
	EventChatMessageAck      = "chat.message.ack"
	EventChatMessageReceive  = "chat.message.receive"
	EventChatMessageFailed   = "chat.message.failed"
	EventChatMessageRecalled = "chat.message.recalled"
)

type Envelope struct {
	Seq       string          `json:"seq"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp int64           `json:"timestamp"`
}

type ErrorData struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	EventType string `json:"event_type,omitempty"`
}

func NewEnvelope(seq string, eventType string, data any) (*Envelope, error) {
	rawData := json.RawMessage("{}")
	if data != nil {
		switch value := data.(type) {
		case json.RawMessage:
			rawData = value
		default:
			marshaled, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			rawData = marshaled
		}
	}

	return &Envelope{
		Seq:       seq,
		Type:      eventType,
		Data:      rawData,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

func MarshalEnvelope(seq string, eventType string, data any) ([]byte, error) {
	envelope, err := NewEnvelope(seq, eventType, data)
	if err != nil {
		return nil, err
	}
	return json.Marshal(envelope)
}
