package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"mini_im/backend/internal/model"
	apperrors "mini_im/backend/internal/pkg/errors"
	"mini_im/backend/internal/pkg/snowflake"
	mysqlrepo "mini_im/backend/internal/repository/mysql"
)

type FriendService struct {
	userRepo         *mysqlrepo.UserRepository
	friendRepo       *mysqlrepo.FriendRepository
	conversationRepo *mysqlrepo.ConversationRepository
	idGenerator      *snowflake.Node
}

func NewFriendService(userRepo *mysqlrepo.UserRepository, friendRepo *mysqlrepo.FriendRepository, conversationRepo *mysqlrepo.ConversationRepository, idGenerator *snowflake.Node) *FriendService {
	return &FriendService{
		userRepo:         userRepo,
		friendRepo:       friendRepo,
		conversationRepo: conversationRepo,
		idGenerator:      idGenerator,
	}
}

func (s *FriendService) CreateFriendRequest(ctx context.Context, fromUserID int64, input CreateFriendRequestInput) (*CreateFriendRequestOutput, *apperrors.AppError) {
	toUserID, appErr := parsePositiveID(input.ToUserID)
	if appErr != nil {
		return nil, appErr
	}
	if fromUserID == toUserID {
		return nil, apperrors.ErrCannotAddSelf
	}

	message := strings.TrimSpace(input.Message)
	if utf8.RuneCountInString(message) > 255 {
		return nil, apperrors.ErrInvalidParam
	}
	if appErr := s.ensureNormalUser(ctx, toUserID); appErr != nil {
		return nil, appErr
	}

	blocked, err := s.friendRepo.IsBlocked(ctx, toUserID, fromUserID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	if blocked {
		return nil, apperrors.ErrBlockedByPeer
	}

	friends, err := s.friendRepo.AreFriends(ctx, fromUserID, toUserID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	if friends {
		return nil, apperrors.ErrAlreadyFriends
	}

	pending, err := s.friendRepo.HasPendingRequestBetween(ctx, fromUserID, toUserID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	if pending {
		return nil, apperrors.ErrFriendRequestPending
	}

	requestID := s.idGenerator.NextID()
	request := &model.FriendRequest{
		RequestID:  requestID,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Message:    sql.NullString{String: message, Valid: message != ""},
		Status:     model.FriendRequestStatusPending,
	}
	if err := s.friendRepo.CreateFriendRequest(ctx, request); err != nil {
		if errors.Is(err, mysqlrepo.ErrFriendRequestPending) {
			return nil, apperrors.ErrFriendRequestPending
		}
		return nil, apperrors.ErrInternal
	}

	return &CreateFriendRequestOutput{
		RequestID: formatID(requestID),
		Status:    model.FriendRequestStatusPending,
	}, nil
}

func (s *FriendService) ListFriendRequests(ctx context.Context, userID int64, direction string, page int, pageSize int) (*PageOutput[FriendRequestOutput], *apperrors.AppError) {
	received := true
	direction = strings.TrimSpace(strings.ToLower(direction))
	switch direction {
	case "", "received":
		received = true
	case "sent":
		received = false
	default:
		return nil, apperrors.ErrInvalidParam
	}

	page, pageSize, offset := normalizePagination(page, pageSize)
	items, total, err := s.friendRepo.ListFriendRequests(ctx, userID, received, pageSize, offset)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	outputs := make([]FriendRequestOutput, 0, len(items))
	for i := range items {
		outputs = append(outputs, toFriendRequestOutput(&items[i]))
	}

	return &PageOutput[FriendRequestOutput]{
		List:     outputs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *FriendService) AcceptFriendRequest(ctx context.Context, userID int64, requestID int64) *apperrors.AppError {
	request, err := s.friendRepo.FindFriendRequest(ctx, requestID)
	if err != nil {
		if errors.Is(err, mysqlrepo.ErrFriendRequestNotFound) {
			return apperrors.ErrFriendRequestNotFound
		}
		return apperrors.ErrInternal
	}
	if request.ToUserID != userID || request.Status != model.FriendRequestStatusPending {
		return apperrors.ErrFriendRequestNotFound
	}

	friends, err := s.friendRepo.AreFriends(ctx, request.FromUserID, request.ToUserID)
	if err != nil {
		return apperrors.ErrInternal
	}
	if friends {
		return apperrors.ErrAlreadyFriends
	}

	var afterAccept func(context.Context, mysqlrepo.Executor) error
	if s.conversationRepo != nil {
		conversationID := s.idGenerator.NextID()
		afterAccept = func(ctx context.Context, exec mysqlrepo.Executor) error {
			_, err := s.conversationRepo.CreateFreshPrivateConversationInTx(ctx, exec, conversationID, request.FromUserID, request.ToUserID)
			return err
		}
	}

	if err := s.friendRepo.AcceptFriendRequest(ctx, requestID, userID, request.FromUserID, request.ToUserID, afterAccept); err != nil {
		if errors.Is(err, mysqlrepo.ErrFriendRequestNotFound) {
			return apperrors.ErrFriendRequestNotFound
		}
		return apperrors.ErrInternal
	}
	return nil
}

func (s *FriendService) RejectFriendRequest(ctx context.Context, userID int64, requestID int64) *apperrors.AppError {
	if err := s.friendRepo.RejectFriendRequest(ctx, requestID, userID); err != nil {
		if errors.Is(err, mysqlrepo.ErrFriendRequestNotFound) {
			return apperrors.ErrFriendRequestNotFound
		}
		return apperrors.ErrInternal
	}
	return nil
}

func (s *FriendService) ListFriends(ctx context.Context, userID int64, page int, pageSize int) (*PageOutput[FriendOutput], *apperrors.AppError) {
	page, pageSize, offset := normalizePagination(page, pageSize)
	items, total, err := s.friendRepo.ListFriends(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	outputs := make([]FriendOutput, 0, len(items))
	for i := range items {
		outputs = append(outputs, FriendOutput{
			FriendUserID:   formatID(items[i].User.User.UserID),
			Nickname:       items[i].User.Profile.Nickname,
			AvatarURL:      items[i].User.Profile.AvatarURL.String,
			Bio:            items[i].User.Profile.Bio.String,
			ConversationID: formatOptionalID(items[i].ConversationID),
			IsBlockedByMe:  items[i].IsBlockedByMe,
			CreatedAt:      items[i].Friendship.CreatedAt,
			UpdatedAt:      items[i].Friendship.UpdatedAt,
		})
	}

	return &PageOutput[FriendOutput]{
		List:     outputs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *FriendService) DeleteFriend(ctx context.Context, userID int64, friendUserID int64) *apperrors.AppError {
	if userID == friendUserID || friendUserID <= 0 {
		return apperrors.ErrInvalidParam
	}

	if err := s.friendRepo.DeleteFriend(ctx, userID, friendUserID); err != nil {
		if errors.Is(err, mysqlrepo.ErrFriendshipNotFound) {
			return apperrors.ErrFriendshipNotFound
		}
		return apperrors.ErrInternal
	}

	return nil
}

func (s *FriendService) BlockUser(ctx context.Context, userID int64, targetUserID int64) *apperrors.AppError {
	if targetUserID <= 0 {
		return apperrors.ErrInvalidParam
	}
	if userID == targetUserID {
		return apperrors.ErrCannotBlockSelf
	}
	if appErr := s.ensureNormalUser(ctx, targetUserID); appErr != nil {
		return appErr
	}

	friends, err := s.friendRepo.AreFriends(ctx, userID, targetUserID)
	if err != nil {
		return apperrors.ErrInternal
	}
	if !friends {
		return apperrors.ErrFriendshipNotFound
	}

	if err := s.friendRepo.BlockUser(ctx, userID, targetUserID); err != nil {
		return apperrors.ErrInternal
	}
	return nil
}

func (s *FriendService) UnblockUser(ctx context.Context, userID int64, targetUserID int64) *apperrors.AppError {
	if targetUserID <= 0 {
		return apperrors.ErrInvalidParam
	}
	if userID == targetUserID {
		return apperrors.ErrCannotBlockSelf
	}

	if err := s.friendRepo.UnblockUser(ctx, userID, targetUserID); err != nil {
		return apperrors.ErrInternal
	}
	return nil
}

func (s *FriendService) ensureNormalUser(ctx context.Context, userID int64) *apperrors.AppError {
	user, err := s.userRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, mysqlrepo.ErrUserNotFound) {
			return apperrors.ErrUserNotFound
		}
		return apperrors.ErrInternal
	}
	if user.User.Status != model.UserStatusNormal {
		return apperrors.ErrUserNotFound
	}
	return nil
}

func parsePositiveID(value string) (int64, *apperrors.AppError) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, apperrors.ErrInvalidParam
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, apperrors.ErrInvalidParam
	}
	return id, nil
}

func formatOptionalID(id sql.NullInt64) string {
	if !id.Valid || id.Int64 <= 0 {
		return ""
	}
	return formatID(id.Int64)
}

func toFriendRequestOutput(item *model.FriendRequestWithUser) FriendRequestOutput {
	return FriendRequestOutput{
		RequestID:  formatID(item.Request.RequestID),
		FromUserID: formatID(item.Request.FromUserID),
		ToUserID:   formatID(item.Request.ToUserID),
		User:       toUserOutput(&item.User),
		Message:    item.Request.Message.String,
		Status:     item.Request.Status,
		CreatedAt:  item.Request.CreatedAt,
		UpdatedAt:  item.Request.UpdatedAt,
	}
}
