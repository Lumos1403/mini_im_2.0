package handler

import (
	"mini_im/backend/internal/middleware"
	apperrors "mini_im/backend/internal/pkg/errors"
	"mini_im/backend/internal/pkg/response"
	"mini_im/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	groupService *service.GroupService
}

func NewGroupHandler(groupService *service.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

func (h *GroupHandler) Register(group *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	groups := group.Group("/groups")
	groups.Use(authMiddleware)
	groups.POST("", h.CreateGroup)
	groups.GET("/search", h.SearchGroups)
	groups.POST("/join-requests/:request_id/accept", h.AcceptJoinRequest)
	groups.POST("/join-requests/:request_id/reject", h.RejectJoinRequest)
	groups.POST("/:group_id/join-requests", h.CreateJoinRequest)
	groups.GET("/:group_id/join-requests", h.ListJoinRequests)
	groups.GET("/:group_id/members", h.ListMembers)
	groups.POST("/:group_id/admins/:user_id", h.SetAdmin)
	groups.DELETE("/:group_id/admins/:user_id", h.UnsetAdmin)
	groups.POST("/:group_id/members/:user_id/mute", h.MuteMember)
	groups.DELETE("/:group_id/members/:user_id/mute", h.UnmuteMember)
	groups.PUT("/:group_id/settings", h.UpdateSettings)
	groups.DELETE("/:group_id", h.DissolveGroup)
}

func (h *GroupHandler) CreateGroup(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	var req service.CreateGroupInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrInvalidParam), apperrors.ErrInvalidParam)
		return
	}
	output, appErr := h.groupService.CreateGroup(ctx.Request.Context(), userID, req)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}
	response.Success(ctx, output)
}

func (h *GroupHandler) SearchGroups(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	output, appErr := h.groupService.SearchGroups(ctx.Request.Context(), userID, ctx.Query("keyword"))
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}
	response.Success(ctx, output)
}

func (h *GroupHandler) CreateJoinRequest(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	var req service.CreateGroupJoinRequestInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrInvalidParam), apperrors.ErrInvalidParam)
		return
	}
	output, appErr := h.groupService.CreateJoinRequest(ctx.Request.Context(), userID, ctx.Param("group_id"), req)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}
	response.Success(ctx, output)
}

func (h *GroupHandler) ListJoinRequests(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	page, pageSize := parsePageQuery(ctx)
	output, appErr := h.groupService.ListJoinRequests(ctx.Request.Context(), userID, ctx.Param("group_id"), page, pageSize)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}
	response.Success(ctx, output)
}

func (h *GroupHandler) AcceptJoinRequest(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	output, appErr := h.groupService.AcceptJoinRequest(ctx.Request.Context(), userID, ctx.Param("request_id"))
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}
	response.Success(ctx, output)
}

func (h *GroupHandler) RejectJoinRequest(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	output, appErr := h.groupService.RejectJoinRequest(ctx.Request.Context(), userID, ctx.Param("request_id"))
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}
	response.Success(ctx, output)
}

func (h *GroupHandler) ListMembers(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	page, pageSize := parsePageQuery(ctx)
	output, appErr := h.groupService.ListMembers(ctx.Request.Context(), userID, ctx.Param("group_id"), page, pageSize)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}
	response.Success(ctx, output)
}

func (h *GroupHandler) SetAdmin(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	if appErr := h.groupService.SetAdmin(ctx.Request.Context(), userID, ctx.Param("group_id"), ctx.Param("user_id"), true); appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}
	response.Success(ctx, gin.H{})
}

func (h *GroupHandler) UnsetAdmin(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	if appErr := h.groupService.SetAdmin(ctx.Request.Context(), userID, ctx.Param("group_id"), ctx.Param("user_id"), false); appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}
	response.Success(ctx, gin.H{})
}

func (h *GroupHandler) MuteMember(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	var req service.MuteGroupMemberInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrInvalidParam), apperrors.ErrInvalidParam)
		return
	}
	if appErr := h.groupService.MuteMember(ctx.Request.Context(), userID, ctx.Param("group_id"), ctx.Param("user_id"), req); appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}
	response.Success(ctx, gin.H{})
}

func (h *GroupHandler) UnmuteMember(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	if appErr := h.groupService.UnmuteMember(ctx.Request.Context(), userID, ctx.Param("group_id"), ctx.Param("user_id")); appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}
	response.Success(ctx, gin.H{})
}

func (h *GroupHandler) UpdateSettings(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	var req service.UpdateGroupSettingsInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrInvalidParam), apperrors.ErrInvalidParam)
		return
	}
	if appErr := h.groupService.UpdateSettings(ctx.Request.Context(), userID, ctx.Param("group_id"), req); appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}
	response.Success(ctx, gin.H{})
}

func (h *GroupHandler) DissolveGroup(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}
	if appErr := h.groupService.DissolveGroup(ctx.Request.Context(), userID, ctx.Param("group_id")); appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}
	response.Success(ctx, gin.H{})
}

func currentUserID(ctx *gin.Context) (int64, bool) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return 0, false
	}
	return userID, true
}
