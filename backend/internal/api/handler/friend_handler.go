package handler

import (
	"strconv"

	"mini_im/backend/internal/middleware"
	apperrors "mini_im/backend/internal/pkg/errors"
	"mini_im/backend/internal/pkg/response"
	"mini_im/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type FriendHandler struct {
	friendService *service.FriendService
}

func NewFriendHandler(friendService *service.FriendService) *FriendHandler {
	return &FriendHandler{friendService: friendService}
}

func (h *FriendHandler) Register(group *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	friends := group.Group("/friends")
	friends.Use(authMiddleware)
	friends.POST("/requests", h.CreateFriendRequest)
	friends.GET("/requests", h.ListFriendRequests)
	friends.POST("/requests/:request_id/accept", h.AcceptFriendRequest)
	friends.POST("/requests/:request_id/reject", h.RejectFriendRequest)
	friends.GET("", h.ListFriends)
	friends.DELETE("/:user_id", h.DeleteFriend)
	friends.POST("/:user_id/block", h.BlockUser)
	friends.DELETE("/:user_id/block", h.UnblockUser)
}

func (h *FriendHandler) CreateFriendRequest(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	var req service.CreateFriendRequestInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrInvalidParam), apperrors.ErrInvalidParam)
		return
	}

	output, appErr := h.friendService.CreateFriendRequest(ctx.Request.Context(), userID, req)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}

func (h *FriendHandler) ListFriendRequests(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	page, pageSize := parsePageQuery(ctx)
	output, appErr := h.friendService.ListFriendRequests(ctx.Request.Context(), userID, ctx.Query("direction"), page, pageSize)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}

func (h *FriendHandler) AcceptFriendRequest(ctx *gin.Context) {
	userID, requestID, ok := h.currentUserAndRequestID(ctx)
	if !ok {
		return
	}

	if appErr := h.friendService.AcceptFriendRequest(ctx.Request.Context(), userID, requestID); appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, gin.H{})
}

func (h *FriendHandler) RejectFriendRequest(ctx *gin.Context) {
	userID, requestID, ok := h.currentUserAndRequestID(ctx)
	if !ok {
		return
	}

	if appErr := h.friendService.RejectFriendRequest(ctx.Request.Context(), userID, requestID); appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, gin.H{})
}

func (h *FriendHandler) ListFriends(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	page, pageSize := parsePageQuery(ctx)
	output, appErr := h.friendService.ListFriends(ctx.Request.Context(), userID, page, pageSize)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}

func (h *FriendHandler) DeleteFriend(ctx *gin.Context) {
	userID, targetUserID, ok := currentUserAndPathUserID(ctx)
	if !ok {
		return
	}

	if appErr := h.friendService.DeleteFriend(ctx.Request.Context(), userID, targetUserID); appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, gin.H{})
}

func (h *FriendHandler) BlockUser(ctx *gin.Context) {
	userID, targetUserID, ok := currentUserAndPathUserID(ctx)
	if !ok {
		return
	}

	if appErr := h.friendService.BlockUser(ctx.Request.Context(), userID, targetUserID); appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, gin.H{})
}

func (h *FriendHandler) UnblockUser(ctx *gin.Context) {
	userID, targetUserID, ok := currentUserAndPathUserID(ctx)
	if !ok {
		return
	}

	if appErr := h.friendService.UnblockUser(ctx.Request.Context(), userID, targetUserID); appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, gin.H{})
}

func (h *FriendHandler) currentUserAndRequestID(ctx *gin.Context) (int64, int64, bool) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return 0, 0, false
	}

	requestID, err := strconv.ParseInt(ctx.Param("request_id"), 10, 64)
	if err != nil || requestID <= 0 {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrInvalidParam), apperrors.ErrInvalidParam)
		return 0, 0, false
	}

	return userID, requestID, true
}

func currentUserAndPathUserID(ctx *gin.Context) (int64, int64, bool) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return 0, 0, false
	}

	targetUserID, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
	if err != nil || targetUserID <= 0 {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrInvalidParam), apperrors.ErrInvalidParam)
		return 0, 0, false
	}

	return userID, targetUserID, true
}

func parsePageQuery(ctx *gin.Context) (int, int) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	return page, pageSize
}
