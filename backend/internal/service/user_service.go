package service

import (
	"context"
	"errors"
	"strings"
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
