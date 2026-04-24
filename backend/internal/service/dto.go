package service

import "time"

type RegisterInput struct {
	Username string
	Password string
	Nickname string
}

type RegisterOutput struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

type LoginInput struct {
	Username string
	Password string
}

type AuthOutput struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	ExpiresIn    int64      `json:"expires_in"`
	User         UserOutput `json:"user,omitempty"`
}

type RefreshOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UserOutput struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

type ProfileOutput struct {
	UserID              string `json:"user_id"`
	Username            string `json:"username"`
	Nickname            string `json:"nickname"`
	AvatarURL           string `json:"avatar_url"`
	Gender              string `json:"gender"`
	Bio                 string `json:"bio"`
	ProfileStatus       string `json:"profile_status"`
	ProfileReviewReason string `json:"profile_review_reason"`
}

type UpdateProfileInput struct {
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Gender    string `json:"gender"`
	Bio       string `json:"bio"`
}

type PageOutput[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type SearchUsersInput struct {
	Keyword  string
	Page     int
	PageSize int
}

type CreateFriendRequestInput struct {
	ToUserID string `json:"to_user_id"`
	Message  string `json:"message"`
}

type FriendRequestOutput struct {
	RequestID  string     `json:"request_id"`
	FromUserID string     `json:"from_user_id"`
	ToUserID   string     `json:"to_user_id"`
	User       UserOutput `json:"user"`
	Message    string     `json:"message"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type CreateFriendRequestOutput struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

type FriendOutput struct {
	User      UserOutput `json:"user"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type ConversationOutput struct {
	ConversationID   string                         `json:"conversation_id"`
	ConversationType string                         `json:"conversation_type"`
	Title            string                         `json:"title"`
	AvatarURL        string                         `json:"avatar_url"`
	LastMessage      *ConversationLastMessageOutput `json:"last_message"`
	UnreadCount      int                            `json:"unread_count"`
	IsPinned         bool                           `json:"is_pinned"`
	IsMuted          bool                           `json:"is_muted"`
}

type ConversationLastMessageOutput struct {
	Content     string    `json:"content"`
	MessageType string    `json:"message_type"`
	CreatedAt   time.Time `json:"created_at"`
}
