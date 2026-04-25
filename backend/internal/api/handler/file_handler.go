package handler

import (
	"mime"
	"net/http"
	"strings"

	"mini_im/backend/internal/middleware"
	apperrors "mini_im/backend/internal/pkg/errors"
	"mini_im/backend/internal/pkg/response"
	"mini_im/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

func (h *FileHandler) Register(group *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	files := group.Group("/files")
	files.Use(authMiddleware)
	files.POST("/upload", h.UploadFile)
	files.GET("/:file_id/download", h.DownloadFile)
}

func (h *FileHandler) UploadFile(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, h.fileService.MaxUploadBytes()+1024*1024)
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		appErr := apperrors.ErrFileInvalid
		if strings.Contains(strings.ToLower(err.Error()), "request body too large") {
			appErr = apperrors.ErrFileTooLarge
		}
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	output, appErr := h.fileService.Upload(ctx.Request.Context(), userID, fileHeader)
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	response.Success(ctx, output)
}

func (h *FileHandler) DownloadFile(ctx *gin.Context) {
	userID, ok := middleware.CurrentUserID(ctx)
	if !ok {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	output, appErr := h.fileService.PrepareDownload(ctx.Request.Context(), userID, ctx.Param("file_id"))
	if appErr != nil {
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}
	defer output.Reader.Close()

	mimeType := strings.TrimSpace(output.MimeType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": output.FileName})
	ctx.DataFromReader(http.StatusOK, output.FileSize, mimeType, output.Reader, map[string]string{
		"Content-Disposition":          disposition,
		"X-Content-Type-Options":       "nosniff",
		"Cache-Control":                "no-store",
		"Content-Security-Policy":      "default-src 'none'",
		"Cross-Origin-Resource-Policy": "same-origin",
	})
}
