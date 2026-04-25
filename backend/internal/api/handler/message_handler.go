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
	conversations.DELETE("/:conversation_id/messages/:message_id", h.DeleteConversationMessage)
	conversations.DELETE("/:conversation_id/messages", h.ClearConversationMessages)

	messages := group.Group("/messages")
	messages.Use(authMiddleware)
	messages.POST("/:message_id/recall", h.RecallMessage)
	messages.GET("/:message_id/recall-edit-cache", h.GetRecallEditCache)
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

func (h *MessageHandler) DeleteConversationMessage(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	if appErr := h.messageService.DeleteConversationMessage(
		ctx.Request.Context(),
		userID,
		ctx.Param("conversation_id"),
		ctx.Param("message_id"),
	); appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, gin.H{})
}

func (h *MessageHandler) ClearConversationMessages(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	if appErr := h.messageService.ClearConversationMessages(
		ctx.Request.Context(),
		userID,
		ctx.Param("conversation_id"),
	); appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, gin.H{})
}

func (h *MessageHandler) RecallMessage(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	output, appErr := h.messageService.RecallMessage(ctx.Request.Context(), userID, ctx.Param("message_id"))
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}

func (h *MessageHandler) GetRecallEditCache(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	output, appErr := h.messageService.GetRecallEditCache(ctx.Request.Context(), userID, ctx.Param("message_id"))
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}
