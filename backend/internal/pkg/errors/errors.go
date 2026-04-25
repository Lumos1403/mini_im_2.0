package errors

import "net/http"

const (
	CodeSuccess = 0

	CodeCommon = 10000
	CodeAuth   = 20000
	CodeUser   = 30000
	CodeFriend = 40000
	CodeMsg    = 50000
	CodeGroup  = 60000
	CodeFile   = 70000
	CodeSystem = 80000
)

type AppError struct {
	Code    int
	Message string
}

func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func (e *AppError) Error() string {
	return e.Message
}

var ErrInternal = New(CodeCommon, "internal error")

var (
	ErrInvalidParam          = New(CodeCommon, "invalid request")
	ErrTokenInvalid          = New(20001, "token invalid")
	ErrTokenExpired          = New(20002, "token expired")
	ErrRefreshTokenInvalid   = New(20003, "refresh token invalid")
	ErrUserNotFound          = New(30001, "user not found")
	ErrUsernameExists        = New(30002, "username already exists")
	ErrInvalidCredentials    = New(30003, "username or password invalid")
	ErrUserDisabled          = New(CodeUser, "user disabled")
	ErrFriendRequestNotFound = New(40001, "friend request not found")
	ErrAlreadyFriends        = New(40002, "already friends")
	ErrBlockedByPeer         = New(40003, "blocked by peer")
	ErrFriendRequestPending  = New(40004, "friend request pending")
	ErrCannotAddSelf         = New(40005, "cannot add yourself")
	ErrFriendshipNotFound    = New(40006, "friendship not found")
	ErrCannotBlockSelf       = New(40007, "cannot block yourself")
	ErrMessageNotFound       = New(50001, "message not found")
	ErrMessageNotRecallable  = New(50002, "message not recallable")
	ErrMessageRecalled       = New(50003, "message already recalled")
	ErrMessageAccessDenied   = New(50004, "message access denied")
	ErrMessageInvalidContent = New(50005, "message content invalid")
	ErrMessageConflict       = New(50006, "client_msg_id already exists with different content")
)

func HTTPStatus(err *AppError) int {
	if err == nil {
		return http.StatusInternalServerError
	}
	if err == ErrInternal {
		return http.StatusInternalServerError
	}

	switch err.Code {
	case CodeCommon:
		return http.StatusBadRequest
	case 20001, 20002, 20003, CodeAuth:
		return http.StatusUnauthorized
	case 30001:
		return http.StatusNotFound
	case 30002, 30003, CodeUser:
		return http.StatusBadRequest
	case 40001, 40006:
		return http.StatusNotFound
	case 40002, 40004:
		return http.StatusConflict
	case 40003:
		return http.StatusForbidden
	case 40005, 40007, CodeFriend:
		return http.StatusBadRequest
	case 50001:
		return http.StatusNotFound
	case 50002:
		return http.StatusBadRequest
	case 50003:
		return http.StatusConflict
	case 50004:
		return http.StatusForbidden
	case 50005:
		return http.StatusBadRequest
	case 50006:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
