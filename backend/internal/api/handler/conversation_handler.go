package handler

import (
	"mini_im/backend/internal/middleware"
	apperrors "mini_im/backend/internal/pkg/errors"
	"mini_im/backend/internal/pkg/response"
	"mini_im/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ConversationHandler struct {
	conversationService *service.ConversationService
}

func NewConversationHandler(conversationService *service.ConversationService) *ConversationHandler {
	return &ConversationHandler{conversationService: conversationService}
}

func (h *ConversationHandler) Register(group *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	conversations := group.Group("/conversations")
	conversations.Use(authMiddleware)
	conversations.GET("", h.ListConversations)
}

func (h *ConversationHandler) ListConversations(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	page, pageSize := parsePageQuery(ctx)
	output, appErr := h.conversationService.ListConversations(ctx.Request.Context(), userID, page, pageSize)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}
