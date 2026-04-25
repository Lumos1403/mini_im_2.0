package model

import (
	"database/sql"
	"time"
)

const (
	GroupRoleOwner  = "owner"
	GroupRoleAdmin  = "admin"
	GroupRoleMember = "member"

	GroupStatusNormal    = "normal"
	GroupStatusDissolved = "dissolved"

	GroupMemberStatusActive = "active"
	GroupMemberStatusLeft   = "left"

	GroupJoinRequestStatusPending  = "pending"
	GroupJoinRequestStatusAccepted = "accepted"
	GroupJoinRequestStatusRejected = "rejected"
)

type Group struct {
	ID                int64
	GroupID           int64
	GroupNo           string
	ConversationID    int64
	OwnerID           int64
	Name              string
	AvatarURL         sql.NullString
	MaxMembers        int
	AllowMemberInvite bool
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type GroupMember struct {
	ID        int64
	GroupID   int64
	UserID    int64
	Role      string
	MuteUntil sql.NullTime
	Status    string
	JoinedAt  time.Time
	LeftAt    sql.NullTime
}

type GroupJoinRequest struct {
	ID        int64
	RequestID int64
	GroupID   int64
	UserID    int64
	Message   sql.NullString
	Status    string
	HandledBy sql.NullInt64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GroupMemberWithProfile struct {
	Member           GroupMember
	User             UserWithProfile
	FriendshipStatus string
}

type GroupJoinRequestWithUser struct {
	Request GroupJoinRequest
	User    UserWithProfile
}

type GroupMessageSender struct {
	UserID    int64
	Nickname  string
	AvatarURL string
	Role      string
}
