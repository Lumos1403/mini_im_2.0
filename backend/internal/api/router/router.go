package router

import (
	"mini_im/backend/internal/api/handler"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	api := engine.Group("/api")
	handler.NewHealthHandler().Register(api)

	return engine
}
