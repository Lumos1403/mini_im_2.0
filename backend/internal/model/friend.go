package model

import (
	"database/sql"
	"time"
)

const (
	FriendRequestStatusPending  = "pending"
	FriendRequestStatusAccepted = "accepted"
	FriendRequestStatusRejected = "rejected"
	FriendRequestStatusExpired  = "expired"

	FriendshipStatusNormal  = "normal"
	FriendshipStatusDeleted = "deleted"
)

type FriendRequest struct {
	ID         int64
	RequestID  int64
	FromUserID int64
	ToUserID   int64
	Message    sql.NullString
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Friendship struct {
	ID        int64
	UserID1   int64
	UserID2   int64
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BlockRelation struct {
	ID        int64
	BlockerID int64
	BlockedID int64
	CreatedAt time.Time
}

type FriendRequestWithUser struct {
	Request FriendRequest
	User    UserWithProfile
}

type FriendshipWithUser struct {
	Friendship Friendship
	User       UserWithProfile
}
