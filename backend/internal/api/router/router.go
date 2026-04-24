package router

import (
	"mini_im/backend/internal/api/handler"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	AuthHandler    *handler.AuthHandler
	UserHandler    *handler.UserHandler
	AuthMiddleware gin.HandlerFunc
}

func New(deps Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	api := engine.Group("/api")
	handler.NewHealthHandler().Register(api)
	if deps.AuthHandler != nil && deps.AuthMiddleware != nil {
		deps.AuthHandler.Register(api, deps.AuthMiddleware)
	}
	if deps.UserHandler != nil && deps.AuthMiddleware != nil {
		deps.UserHandler.Register(api, deps.AuthMiddleware)
	}

	return engine
}
