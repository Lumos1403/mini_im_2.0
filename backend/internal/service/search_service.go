package service

import (
	"context"
	"strings"

	"mini_im/backend/internal/model"
	apperrors "mini_im/backend/internal/pkg/errors"
	mysqlrepo "mini_im/backend/internal/repository/mysql"
)

type SearchService struct {
	searchRepo *mysqlrepo.SearchRepository
	userRepo   *mysqlrepo.UserRepository
}

func NewSearchService(searchRepo *mysqlrepo.SearchRepository, userRepo *mysqlrepo.UserRepository) *SearchService {
	return &SearchService{searchRepo: searchRepo, userRepo: userRepo}
}

func (s *SearchService) SearchMessages(ctx context.Context, userID int64, keyword string, page int, pageSize int) (*PageOutput[SearchMessageItem], *apperrors.AppError) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, apperrors.ErrInvalidParam
	}

	page, pageSize, offset := normalizePagination(page, pageSize)
	rows, total, err := s.searchRepo.SearchMessages(ctx, userID, keyword, pageSize, offset)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	senderProfiles := s.batchQueryProfiles(ctx, uniqueSearchSenderIDs(rows))
	items := make([]SearchMessageItem, 0, len(rows))
	for _, r := range rows {
		item := SearchMessageItem{
			MessageID:        formatID(r.MessageID),
			ConversationID:   formatID(r.ConversationID),
			ConversationType: r.ConversationType,
			SenderID:         formatID(r.SenderID),
			MessageType:      r.MessageType,
			Content:          r.Content.String,
			CreatedAt:        formatTime(r.CreatedAt),
		}
		if p, ok := senderProfiles[r.SenderID]; ok {
			item.SenderNickname = p.Nickname
			item.SenderAvatarURL = p.AvatarURL
		}
		items = append(items, item)
	}

	return &PageOutput[SearchMessageItem]{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *SearchService) SearchFiles(ctx context.Context, userID int64, keyword string, page int, pageSize int) (*PageOutput[SearchFileItem], *apperrors.AppError) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, apperrors.ErrInvalidParam
	}

	page, pageSize, offset := normalizePagination(page, pageSize)
	rows, total, err := s.searchRepo.SearchFiles(ctx, userID, keyword, pageSize, offset)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	uploaderProfiles := s.batchQueryProfiles(ctx, uniqueSearchFileUploaderIDs(rows))
	items := make([]SearchFileItem, 0, len(rows))
	for _, r := range rows {
		item := SearchFileItem{
			FileID:           formatID(r.FileID),
			OriginalName:     r.OriginalName,
			FileSize:         r.FileSize,
			MimeType:         r.MimeType.String,
			UploaderID:       formatID(r.UploaderID),
			MessageID:        formatID(r.MessageID),
			ConversationID:   formatID(r.ConversationID),
			ConversationType: r.ConversationType,
			CreatedAt:        formatTime(r.CreatedAt),
		}
		if p, ok := uploaderProfiles[r.UploaderID]; ok {
			item.UploaderNickname = p.Nickname
		}
		items = append(items, item)
	}

	return &PageOutput[SearchFileItem]{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *SearchService) batchQueryProfiles(ctx context.Context, userIDs []int64) map[int64]userProfileResult {
	result := make(map[int64]userProfileResult, len(userIDs))
	if s.userRepo == nil || len(userIDs) == 0 {
		return result
	}
	for _, userID := range userIDs {
		user, err := s.userRepo.FindByUserID(ctx, userID)
		if err != nil || user == nil {
			continue
		}
		result[userID] = toUserProfileResult(&user.Profile)
	}
	return result
}

type userProfileResult struct {
	Nickname  string
	AvatarURL string
}

func toUserProfileResult(profile *model.UserProfile) userProfileResult {
	return userProfileResult{
		Nickname:  profile.Nickname,
		AvatarURL: profile.AvatarURL.String,
	}
}

func uniqueSearchSenderIDs(rows []mysqlrepo.MessageSearchRow) []int64 {
	seen := make(map[int64]struct{})
	ids := make([]int64, 0)
	for _, r := range rows {
		if r.SenderID <= 0 {
			continue
		}
		if _, ok := seen[r.SenderID]; ok {
			continue
		}
		seen[r.SenderID] = struct{}{}
		ids = append(ids, r.SenderID)
	}
	return ids
}

func uniqueSearchFileUploaderIDs(rows []mysqlrepo.FileSearchRow) []int64 {
	seen := make(map[int64]struct{})
	ids := make([]int64, 0)
	for _, r := range rows {
		if r.UploaderID <= 0 {
			continue
		}
		if _, ok := seen[r.UploaderID]; ok {
			continue
		}
		seen[r.UploaderID] = struct{}{}
		ids = append(ids, r.UploaderID)
	}
	return ids
}
