package service

import (
	"encoding/json"
	"time"
)

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

type UserSearchOutput struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
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
	FriendUserID   string    `json:"friend_user_id"`
	Nickname       string    `json:"nickname"`
	AvatarURL      string    `json:"avatar_url"`
	Bio            string    `json:"bio"`
	ConversationID string    `json:"conversation_id"`
	IsBlockedByMe  bool      `json:"is_blocked_by_me"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ConversationOutput struct {
	ConversationID   string                         `json:"conversation_id"`
	ConversationType string                         `json:"conversation_type"`
	Title            string                         `json:"title"`
	AvatarURL        string                         `json:"avatar_url"`
	PeerUserID       string                         `json:"peer_user_id"`
	PeerNickname     string                         `json:"peer_nickname"`
	PeerAvatarURL    string                         `json:"peer_avatar_url"`
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

type SendMessageAckOutput struct {
	ClientMsgID    string `json:"client_msg_id"`
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	SendStatus     string `json:"send_status"`
	ServerTime     string `json:"server_time"`
}

type SendMessageFailedOutput struct {
	ClientMsgID    string `json:"client_msg_id"`
	MessageID      string `json:"message_id,omitempty"`
	ConversationID string `json:"conversation_id"`
	SendStatus     string `json:"send_status"`
	Code           string `json:"code"`
	Message        string `json:"message"`
	ServerTime     string `json:"server_time,omitempty"`
}

type MessageReceiveOutput struct {
	ClientMsgID    string          `json:"client_msg_id"`
	MessageID      string          `json:"message_id"`
	ConversationID string          `json:"conversation_id"`
	SenderID       string          `json:"sender_id"`
	MessageType    string          `json:"message_type"`
	Content        string          `json:"content"`
	ExtraJSON      json.RawMessage `json:"extra_json"`
	SendStatus     string          `json:"send_status"`
	CreatedAt      string          `json:"created_at"`
}

type MessageOutput struct {
	ClientMsgID    string          `json:"client_msg_id"`
	MessageID      string          `json:"message_id"`
	ConversationID string          `json:"conversation_id"`
	SenderID       string          `json:"sender_id"`
	MessageType    string          `json:"message_type"`
	Content        string          `json:"content"`
	ExtraJSON      json.RawMessage `json:"extra_json"`
	SendStatus     string          `json:"send_status"`
	CreatedAt      string          `json:"created_at"`
}

type MessagePageOutput struct {
	List       []MessageOutput `json:"list"`
	NextCursor string          `json:"next_cursor"`
	HasMore    bool            `json:"has_more"`
	Limit      int             `json:"limit"`
}

type RecallMessageOutput struct {
	MessageID     string `json:"message_id"`
	EditableUntil string `json:"editable_until"`
}

type RecallEditCacheOutput struct {
	MessageID string `json:"message_id"`
	Content   string `json:"content"`
}

type MessageRecalledEventOutput struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	RecalledBy     string `json:"recalled_by"`
	RecalledAt     string `json:"recalled_at"`
}

type FileUploadOutput struct {
	FileID       string `json:"file_id"`
	OriginalName string `json:"original_name"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
}
