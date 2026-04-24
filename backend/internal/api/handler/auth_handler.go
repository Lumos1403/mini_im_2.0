package handler

import (
	"mini_im/backend/internal/middleware"
	apperrors "mini_im/backend/internal/pkg/errors"
	"mini_im/backend/internal/pkg/response"
	"mini_im/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(group *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	auth := group.Group("/auth")
	auth.POST("/register", h.RegisterUser)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/logout", authMiddleware, h.Logout)
}

func (h *AuthHandler) RegisterUser(ctx *gin.Context) {
	var req service.RegisterInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrInvalidParam), apperrors.ErrInvalidParam)
		return
	}

	output, appErr := h.authService.Register(ctx.Request.Context(), req)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var req service.LoginInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrInvalidParam), apperrors.ErrInvalidParam)
		return
	}

	output, appErr := h.authService.Login(ctx.Request.Context(), req)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}

func (h *AuthHandler) Refresh(ctx *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrInvalidParam), apperrors.ErrInvalidParam)
		return
	}

	output, appErr := h.authService.Refresh(ctx.Request.Context(), req.RefreshToken)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}

func (h *AuthHandler) Logout(ctx *gin.Context) {
	userID, okUser := middleware.CurrentUserID(ctx)
	deviceID, okDevice := middleware.CurrentDeviceID(ctx)
	jti, okJTI := middleware.CurrentAccessJTI(ctx)
	expiresAt, okExpiresAt := middleware.CurrentAccessExpiresAt(ctx)
	if !okUser || !okDevice || !okJTI || !okExpiresAt {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	if appErr := h.authService.Logout(ctx.Request.Context(), userID, deviceID, jti, expiresAt); appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, gin.H{})
}
