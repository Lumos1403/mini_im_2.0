package model

import (
	"database/sql"
	"time"
)

const (
	MessageTypeText   = "text"
	MessageTypeEmoji  = "emoji"
	MessageTypeFile   = "file"
	MessageTypeSystem = "system"

	MessageSendStatusSent          = "sent"
	MessageSendStatusFailed        = "failed"
	MessageSendStatusFailedBlocked = "failed_blocked"
	MessageSendStatusRecalled      = "recalled"
)

type Message struct {
	ID             int64
	MessageID      int64
	ConversationID int64
	SenderID       int64
	ClientMsgID    string
	MessageType    string
	Content        sql.NullString
	ExtraJSON      sql.NullString
	SendStatus     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	RecalledAt     sql.NullTime
	RecalledBy     sql.NullInt64
	IsDeletedAll   bool
}

type MessageUserState struct {
	ID        int64
	MessageID int64
	UserID    int64
	IsDeleted bool
	DeletedAt sql.NullTime
	CreatedAt time.Time
	UpdatedAt time.Time
}
