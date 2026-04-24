package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"mini_im/backend/internal/model"
	apperrors "mini_im/backend/internal/pkg/errors"
	jwtpkg "mini_im/backend/internal/pkg/jwt"
	"mini_im/backend/internal/pkg/password"
	"mini_im/backend/internal/pkg/snowflake"
	mysqlrepo "mini_im/backend/internal/repository/mysql"
	redisrepo "mini_im/backend/internal/repository/redis"
)

type AuthService struct {
	userRepo     *mysqlrepo.UserRepository
	tokenRepo    *redisrepo.TokenRepository
	tokenManager *jwtpkg.Manager
	idGenerator  *snowflake.Node
	agentService *AgentService
}

func NewAuthService(userRepo *mysqlrepo.UserRepository, tokenRepo *redisrepo.TokenRepository, tokenManager *jwtpkg.Manager, idGenerator *snowflake.Node, agentService *AgentService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		tokenManager: tokenManager,
		idGenerator:  idGenerator,
		agentService: agentService,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, *apperrors.AppError) {
	username := strings.TrimSpace(input.Username)
	nickname := strings.TrimSpace(input.Nickname)
	if !validUsername(username) || !validPassword(input.Password) || nickname == "" || len([]rune(nickname)) > 64 {
		return nil, apperrors.ErrInvalidParam
	}

	if _, err := s.userRepo.FindByUsername(ctx, username); err == nil {
		return nil, apperrors.ErrUsernameExists
	} else if !errors.Is(err, mysqlrepo.ErrUserNotFound) {
		return nil, apperrors.ErrInternal
	}

	passwordHash, err := password.Hash(input.Password)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	userID := s.idGenerator.NextID()
	user := &model.User{
		UserID:       userID,
		Username:     username,
		PasswordHash: passwordHash,
		UserType:     model.UserTypeNormal,
		Status:       model.UserStatusNormal,
	}
	profile := &model.UserProfile{
		UserID:        userID,
		Nickname:      nickname,
		ProfileStatus: model.UserStatusNormal,
	}

	if err := s.userRepo.CreateUserWithProfile(ctx, user, profile); err != nil {
		if errors.Is(err, mysqlrepo.ErrDuplicateUser) {
			return nil, apperrors.ErrUsernameExists
		}
		return nil, apperrors.ErrInternal
	}

	if err := s.agentService.EnsureDefaultAgentFriend(ctx, userID); err != nil {
		return nil, apperrors.ErrInternal
	}

	return &RegisterOutput{
		UserID:   formatID(userID),
		Username: username,
		Nickname: nickname,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthOutput, *apperrors.AppError) {
	username := strings.TrimSpace(input.Username)
	if username == "" || input.Password == "" {
		return nil, apperrors.ErrInvalidParam
	}

	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, mysqlrepo.ErrUserNotFound) {
			return nil, apperrors.ErrInvalidCredentials
		}
		return nil, apperrors.ErrInternal
	}
	if user.User.Status != model.UserStatusNormal {
		return nil, apperrors.ErrUserDisabled
	}
	if !password.Compare(user.User.PasswordHash, input.Password) {
		return nil, apperrors.ErrInvalidCredentials
	}

	deviceID, err := jwtpkg.RandomHex(16)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	pair, err := s.tokenManager.GeneratePair(user.User.UserID, deviceID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	if err := s.tokenRepo.SaveRefreshToken(ctx, user.User.UserID, deviceID, pair.RefreshToken.Claims.JTI, s.tokenManager.RefreshTTL()); err != nil {
		return nil, apperrors.ErrInternal
	}

	return &AuthOutput{
		AccessToken:  pair.AccessToken.Value,
		RefreshToken: pair.RefreshToken.Value,
		ExpiresIn:    pair.ExpiresIn,
		User:         toUserOutput(user),
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*RefreshOutput, *apperrors.AppError) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, apperrors.ErrRefreshTokenInvalid
	}

	claims, err := s.tokenManager.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, apperrors.ErrRefreshTokenInvalid
	}

	storedJTI, err := s.tokenRepo.GetRefreshTokenJTI(ctx, claims.UserID, claims.DeviceID)
	if err != nil {
		return nil, apperrors.ErrRefreshTokenInvalid
	}
	if storedJTI != claims.JTI {
		return nil, apperrors.ErrRefreshTokenInvalid
	}

	pair, err := s.tokenManager.GeneratePair(claims.UserID, claims.DeviceID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	if err := s.tokenRepo.SaveRefreshToken(ctx, claims.UserID, claims.DeviceID, pair.RefreshToken.Claims.JTI, s.tokenManager.RefreshTTL()); err != nil {
		return nil, apperrors.ErrInternal
	}

	return &RefreshOutput{
		AccessToken:  pair.AccessToken.Value,
		RefreshToken: pair.RefreshToken.Value,
		ExpiresIn:    pair.ExpiresIn,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID int64, deviceID string, accessJTI string, accessExpiresAt int64) *apperrors.AppError {
	if userID <= 0 || deviceID == "" || accessJTI == "" {
		return apperrors.ErrTokenInvalid
	}
	if err := s.tokenRepo.DeleteRefreshToken(ctx, userID, deviceID); err != nil {
		return apperrors.ErrInternal
	}

	ttl := time.Until(time.Unix(accessExpiresAt, 0))
	if err := s.tokenRepo.BlacklistAccessToken(ctx, accessJTI, ttl); err != nil {
		return apperrors.ErrInternal
	}

	return nil
}

func validUsername(username string) bool {
	length := len(username)
	return length >= 3 && length <= 64
}

func validPassword(raw string) bool {
	length := len(raw)
	return length >= 6 && length <= 72
}

func toUserOutput(user *model.UserWithProfile) UserOutput {
	return UserOutput{
		UserID:    formatID(user.User.UserID),
		Username:  user.User.Username,
		Nickname:  user.Profile.Nickname,
		AvatarURL: user.Profile.AvatarURL.String,
	}
}

func toProfileOutput(user *model.UserWithProfile) ProfileOutput {
	return ProfileOutput{
		UserID:              formatID(user.User.UserID),
		Username:            user.User.Username,
		Nickname:            user.Profile.Nickname,
		AvatarURL:           user.Profile.AvatarURL.String,
		Gender:              user.Profile.Gender.String,
		Bio:                 user.Profile.Bio.String,
		ProfileStatus:       user.Profile.ProfileStatus,
		ProfileReviewReason: user.Profile.ProfileReviewReason.String,
	}
}

func formatID(id int64) string {
	return strconv.FormatInt(id, 10)
}
