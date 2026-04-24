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
	users.GET("/me", h.Me)
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
