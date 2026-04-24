package handler

import (
	"mini_im/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Register(group *gin.RouterGroup) {
	group.GET("/health", h.Health)
}

func (h *HealthHandler) Health(ctx *gin.Context) {
	response.Success(ctx, gin.H{
		"status": "ok",
	})
}
