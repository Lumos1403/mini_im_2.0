package service

import (
	"context"
	"errors"

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
