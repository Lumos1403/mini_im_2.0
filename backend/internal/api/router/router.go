package router

import (
	"mini_im/backend/internal/api/handler"
	"mini_im/backend/internal/ws"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	AuthHandler         *handler.AuthHandler
	UserHandler         *handler.UserHandler
	FriendHandler       *handler.FriendHandler
	GroupHandler        *handler.GroupHandler
	ConversationHandler *handler.ConversationHandler
	MessageHandler      *handler.MessageHandler
	FileHandler         *handler.FileHandler
	SearchHandler       *handler.SearchHandler
	AuthMiddleware      gin.HandlerFunc
	WSServer            *ws.Server
}

func New(deps Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	if deps.WSServer != nil {
		engine.GET("/ws", deps.WSServer.Handle)
	}

	api := engine.Group("/api")
	handler.NewHealthHandler().Register(api)
	if deps.AuthHandler != nil && deps.AuthMiddleware != nil {
		deps.AuthHandler.Register(api, deps.AuthMiddleware)
	}
	if deps.UserHandler != nil && deps.AuthMiddleware != nil {
		deps.UserHandler.Register(api, deps.AuthMiddleware)
	}
	if deps.FriendHandler != nil && deps.AuthMiddleware != nil {
		deps.FriendHandler.Register(api, deps.AuthMiddleware)
	}
	if deps.GroupHandler != nil && deps.AuthMiddleware != nil {
		deps.GroupHandler.Register(api, deps.AuthMiddleware)
	}
	if deps.ConversationHandler != nil && deps.AuthMiddleware != nil {
		deps.ConversationHandler.Register(api, deps.AuthMiddleware)
	}
	if deps.MessageHandler != nil && deps.AuthMiddleware != nil {
		deps.MessageHandler.Register(api, deps.AuthMiddleware)
	}
	if deps.FileHandler != nil && deps.AuthMiddleware != nil {
		deps.FileHandler.Register(api, deps.AuthMiddleware)
	}
	if deps.SearchHandler != nil && deps.AuthMiddleware != nil {
		deps.SearchHandler.Register(api, deps.AuthMiddleware)
	}

	return engine
}
