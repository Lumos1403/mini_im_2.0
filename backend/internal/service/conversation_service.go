package service

import (
	"context"

	"mini_im/backend/internal/model"
	apperrors "mini_im/backend/internal/pkg/errors"
	mysqlrepo "mini_im/backend/internal/repository/mysql"
)

type ConversationService struct {
	conversationRepo *mysqlrepo.ConversationRepository
}

func NewConversationService(conversationRepo *mysqlrepo.ConversationRepository) *ConversationService {
	return &ConversationService{conversationRepo: conversationRepo}
}

func (s *ConversationService) ListConversations(ctx context.Context, userID int64, page int, pageSize int) (*PageOutput[ConversationOutput], *apperrors.AppError) {
	page, pageSize, offset := normalizePagination(page, pageSize)
	items, total, err := s.conversationRepo.ListUserConversations(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	outputs := make([]ConversationOutput, 0, len(items))
	for i := range items {
		outputs = append(outputs, toConversationOutput(&items[i]))
	}

	return &PageOutput[ConversationOutput]{
		List:     outputs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func toConversationOutput(item *model.ConversationListItem) ConversationOutput {
	output := ConversationOutput{
		ConversationID:   formatID(item.ConversationID),
		ConversationType: item.ConversationType,
		UnreadCount:      item.UnreadCount,
		IsPinned:         item.IsPinned,
		IsMuted:          item.IsMuted,
	}

	if item.ConversationType == model.ConversationTypePrivate && item.Peer.User.UserID > 0 {
		output.Title = item.Peer.Profile.Nickname
		output.AvatarURL = item.Peer.Profile.AvatarURL.String
	}

	return output
}
