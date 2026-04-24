package handler

import (
	"mini_im/backend/internal/middleware"
	apperrors "mini_im/backend/internal/pkg/errors"
	"mini_im/backend/internal/pkg/response"
	"mini_im/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(group *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	users := group.Group("/users")
	users.Use(authMiddleware)
	users.GET("/search", h.Search)
	users.GET("/me", h.Me)
	users.GET("/me/profile", h.Profile)
	users.PUT("/me/profile", h.UpdateProfile)
}

func (h *UserHandler) Search(ctx *gin.Context) {
	if _, ok := middleware.CurrentUserID(ctx); !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	page, pageSize := parsePageQuery(ctx)
	output, appErr := h.userService.SearchUsers(ctx.Request.Context(), service.SearchUsersInput{
		Keyword:  ctx.Query("keyword"),
		Page:     page,
		PageSize: pageSize,
	})
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}

func (h *UserHandler) Me(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	output, appErr := h.userService.GetMe(ctx.Request.Context(), userID)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}

func (h *UserHandler) Profile(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	output, appErr := h.userService.GetProfile(ctx.Request.Context(), userID)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}

func (h *UserHandler) UpdateProfile(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	var req service.UpdateProfileInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrInvalidParam), apperrors.ErrInvalidParam)
		return
	}

	output, appErr := h.userService.UpdateProfile(ctx.Request.Context(), userID, req)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}
