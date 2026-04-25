package model

import (
	"database/sql"
	"time"
)

const (
	ConversationTypePrivate = "private"
	ConversationTypeGroup   = "group"

	ConversationStatusNormal = "normal"

	ConversationMemberRoleOwner  = "owner"
	ConversationMemberRoleAdmin  = "admin"
	ConversationMemberRoleMember = "member"

	ConversationMemberStatusActive = "active"
	ConversationMemberStatusLeft   = "left"
)

type Conversation struct {
	ID               int64
	ConversationID   int64
	ConversationType string
	RefID            sql.NullInt64
	LastMessageID    sql.NullInt64
	LastMessageAt    sql.NullTime
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ConversationMember struct {
	ID             int64
	ConversationID int64
	UserID         int64
	Role           string
	Status         string
	JoinedAt       time.Time
	LeftAt         sql.NullTime
}

type ConversationUserState struct {
	ID                int64
	ConversationID    int64
	UserID            int64
	IsDeleted         bool
	ClearedAt         sql.NullTime
	LastReadMessageID sql.NullInt64
	LastReadAt        sql.NullTime
	UnreadCount       int
	IsPinned          bool
	IsMuted           bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type ConversationListItem struct {
	ConversationID   int64
	ConversationType string
	LastMessageID    sql.NullInt64
	LastMessageAt    sql.NullTime
	UnreadCount      int
	IsPinned         bool
	IsMuted          bool
	UpdatedAt        time.Time
	Peer             UserWithProfile
	GroupID          sql.NullInt64
	GroupNo          sql.NullString
	GroupName        sql.NullString
	GroupAvatarURL   sql.NullString
	GroupStatus      sql.NullString
}
