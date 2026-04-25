package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	apperrors "mini_im/backend/internal/pkg/errors"
	mysqlrepo "mini_im/backend/internal/repository/mysql"
)

type UserService struct {
	userRepo *mysqlrepo.UserRepository
}

func NewUserService(userRepo *mysqlrepo.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetMe(ctx context.Context, userID int64) (*UserOutput, *apperrors.AppError) {
	user, err := s.userRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, mysqlrepo.ErrUserNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.ErrInternal
	}

	output := toUserOutput(user)
	return &output, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID int64) (*ProfileOutput, *apperrors.AppError) {
	user, err := s.userRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, mysqlrepo.ErrUserNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.ErrInternal
	}

	output := toProfileOutput(user)
	return &output, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, input UpdateProfileInput) (*ProfileOutput, *apperrors.AppError) {
	profile, appErr := normalizeProfileInput(input)
	if appErr != nil {
		return nil, appErr
	}

	if err := s.userRepo.UpdateProfile(ctx, userID, mysqlrepo.UpdateProfileParams{
		Nickname:  profile.Nickname,
		AvatarURL: profile.AvatarURL,
		Gender:    profile.Gender,
		Bio:       profile.Bio,
	}); err != nil {
		return nil, apperrors.ErrInternal
	}

	return s.GetProfile(ctx, userID)
}

func (s *UserService) SearchUsers(ctx context.Context, input SearchUsersInput) (*PageOutput[UserSearchOutput], *apperrors.AppError) {
	keyword := strings.TrimSpace(input.Keyword)
	if keyword == "" {
		return nil, apperrors.ErrInvalidParam
	}

	page, pageSize, offset := normalizePagination(input.Page, input.PageSize)
	params := mysqlrepo.SearchUsersParams{
		Keyword: keyword,
		Limit:   pageSize,
		Offset:  offset,
	}
	if isDigits(keyword) {
		userID, err := strconv.ParseInt(keyword, 10, 64)
		if err != nil || userID <= 0 {
			return nil, apperrors.ErrInvalidParam
		}
		params.ByUserID = true
		params.UserID = userID
	}

	users, total, err := s.userRepo.Search(ctx, params)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	outputs := make([]UserSearchOutput, 0, len(users))
	for i := range users {
		outputs = append(outputs, toUserSearchOutput(&users[i]))
	}

	return &PageOutput[UserSearchOutput]{
		List:     outputs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func normalizeProfileInput(input UpdateProfileInput) (UpdateProfileInput, *apperrors.AppError) {
	profile := UpdateProfileInput{
		Nickname:  strings.TrimSpace(input.Nickname),
		AvatarURL: strings.TrimSpace(input.AvatarURL),
		Gender:    strings.TrimSpace(input.Gender),
		Bio:       strings.TrimSpace(input.Bio),
	}

	if profile.Nickname == "" ||
		utf8.RuneCountInString(profile.Nickname) > 64 ||
		utf8.RuneCountInString(profile.AvatarURL) > 512 ||
		utf8.RuneCountInString(profile.Gender) > 20 ||
		utf8.RuneCountInString(profile.Bio) > 255 {
		return UpdateProfileInput{}, apperrors.ErrInvalidParam
	}

	return profile, nil
}

func normalizePagination(page int, pageSize int) (int, int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize, (page - 1) * pageSize
}

func isDigits(value string) bool {
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return value != ""
}
