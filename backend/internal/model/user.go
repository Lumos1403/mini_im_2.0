package model

import (
	"database/sql"
	"time"
)

const (
	UserTypeNormal = "normal"
	UserTypeAgent  = "agent"
	UserTypeSystem = "system"

	UserStatusNormal   = "normal"
	UserStatusDisabled = "disabled"
	UserStatusDeleted  = "deleted"
)

type User struct {
	ID           int64
	UserID       int64
	Username     string
	PasswordHash string
	UserType     string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    sql.NullTime
}

type UserProfile struct {
	ID                  int64
	UserID              int64
	Nickname            string
	AvatarURL           sql.NullString
	Gender              sql.NullString
	Bio                 sql.NullString
	ProfileStatus       string
	ProfileReviewReason sql.NullString
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type UserWithProfile struct {
	User    User
	Profile UserProfile
}
