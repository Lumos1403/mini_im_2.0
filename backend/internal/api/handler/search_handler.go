package handler

import (
	"mini_im/backend/internal/middleware"
	apperrors "mini_im/backend/internal/pkg/errors"
	"mini_im/backend/internal/pkg/response"
	"mini_im/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

func (h *SearchHandler) Register(group *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	search := group.Group("/search")
	search.Use(authMiddleware)
	search.GET("/messages", h.SearchMessages)
	search.GET("/files", h.SearchFiles)
}

func (h *SearchHandler) SearchMessages(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	keyword := ctx.Query("keyword")
	page, pageSize := parsePageQuery(ctx)

	output, appErr := h.searchService.SearchMessages(ctx.Request.Context(), userID, keyword, page, pageSize)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}

func (h *SearchHandler) SearchFiles(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	keyword := ctx.Query("keyword")
	page, pageSize := parsePageQuery(ctx)

	output, appErr := h.searchService.SearchFiles(ctx.Request.Context(), userID, keyword, page, pageSize)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}
