package middleware

import (
	"errors"
	"strings"

	apperrors "mini_im/backend/internal/pkg/errors"
	jwtpkg "mini_im/backend/internal/pkg/jwt"
	"mini_im/backend/internal/pkg/response"
	redisrepo "mini_im/backend/internal/repository/redis"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey          = "auth_user_id"
	ContextDeviceIDKey        = "auth_device_id"
	ContextAccessJTIKey       = "auth_access_jti"
	ContextAccessExpiresAtKey = "auth_access_expires_at"
)

func NewAuthMiddleware(tokenManager *jwtpkg.Manager, tokenRepo *redisrepo.TokenRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
			ctx.Abort()
			return
		}

		tokenValue := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		claims, err := tokenManager.ParseAccessToken(tokenValue)
		if err != nil {
			appErr := apperrors.ErrTokenInvalid
			if errors.Is(err, jwtpkg.ErrExpiredToken) {
				appErr = apperrors.ErrTokenExpired
			}
			response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
			ctx.Abort()
			return
		}

		blacklisted, err := tokenRepo.IsAccessTokenBlacklisted(ctx.Request.Context(), claims.JTI)
		if err != nil {
			response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrInternal), apperrors.ErrInternal)
			ctx.Abort()
			return
		}
		if blacklisted {
			response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
			ctx.Abort()
			return
		}

		ctx.Set(ContextUserIDKey, claims.UserID)
		ctx.Set(ContextDeviceIDKey, claims.DeviceID)
		ctx.Set(ContextAccessJTIKey, claims.JTI)
		ctx.Set(ContextAccessExpiresAtKey, claims.ExpiresAt)
		ctx.Next()
	}
}

func CurrentUserID(ctx *gin.Context) (int64, bool) {
	value, ok := ctx.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	userID, ok := value.(int64)
	return userID, ok
}

func CurrentDeviceID(ctx *gin.Context) (string, bool) {
	value, ok := ctx.Get(ContextDeviceIDKey)
	if !ok {
		return "", false
	}
	deviceID, ok := value.(string)
	return deviceID, ok
}

func CurrentAccessJTI(ctx *gin.Context) (string, bool) {
	value, ok := ctx.Get(ContextAccessJTIKey)
	if !ok {
		return "", false
	}
	jti, ok := value.(string)
	return jti, ok
}

func CurrentAccessExpiresAt(ctx *gin.Context) (int64, bool) {
	value, ok := ctx.Get(ContextAccessExpiresAtKey)
	if !ok {
		return 0, false
	}
	expiresAt, ok := value.(int64)
	return expiresAt, ok
}
