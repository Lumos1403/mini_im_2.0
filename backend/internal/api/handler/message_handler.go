package handler

import (
	"strconv"

	"mini_im/backend/internal/middleware"
	apperrors "mini_im/backend/internal/pkg/errors"
	"mini_im/backend/internal/pkg/response"
	"mini_im/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

func (h *MessageHandler) Register(group *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	conversations := group.Group("/conversations")
	conversations.Use(authMiddleware)
	conversations.GET("/:conversation_id/messages", h.ListConversationMessages)
}

func (h *MessageHandler) ListConversationMessages(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "30"))
	output, appErr := h.messageService.ListConversationMessages(
		ctx.Request.Context(),
		userID,
		ctx.Param("conversation_id"),
		ctx.Query("cursor"),
		limit,
	)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}
