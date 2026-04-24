package main

import (
	"context"
	"time"

	"mini_im/backend/internal/api/handler"
	"mini_im/backend/internal/api/router"
	"mini_im/backend/internal/config"
	"mini_im/backend/internal/middleware"
	jwtpkg "mini_im/backend/internal/pkg/jwt"
	"mini_im/backend/internal/pkg/logger"
	"mini_im/backend/internal/pkg/snowflake"
	mysqlrepo "mini_im/backend/internal/repository/mysql"
	redisrepo "mini_im/backend/internal/repository/redis"
	"mini_im/backend/internal/service"
	"mini_im/backend/internal/ws"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	gin.SetMode(cfg.Server.Mode)
	logger.Init(cfg.Server.Mode)
	defer logger.Sync()

	db, err := mysqlrepo.New(cfg.MySQL)
	if err != nil {
		logger.L().Fatal("mysql init failed", zap.Error(err))
	}
	defer db.Close()

	redisClient, err := redisrepo.New(cfg.Redis)
	if err != nil {
		logger.L().Fatal("redis init failed", zap.Error(err))
	}
	defer redisClient.Close()

	idGenerator, err := snowflake.NewNode(cfg.Snowflake.NodeID)
	if err != nil {
		logger.L().Fatal("snowflake init failed", zap.Error(err))
	}

	tokenManager, err := jwtpkg.NewManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		time.Duration(cfg.JWT.AccessExpireMinutes)*time.Minute,
		time.Duration(cfg.JWT.RefreshExpireDays)*24*time.Hour,
	)
	if err != nil {
		logger.L().Fatal("jwt init failed", zap.Error(err))
	}

	userRepo := mysqlrepo.NewUserRepository(db)
	friendRepo := mysqlrepo.NewFriendRepository(db)
	conversationRepo := mysqlrepo.NewConversationRepository(db)
	messageRepo := mysqlrepo.NewMessageRepository(db)
	tokenRepo := redisrepo.NewTokenRepository(redisClient)
	onlineRepo := redisrepo.NewOnlineRepository(redisClient)
	agentService := service.NewAgentService()
	authService := service.NewAuthService(userRepo, tokenRepo, tokenManager, idGenerator, agentService)
	userService := service.NewUserService(userRepo)
	conversationService := service.NewConversationService(conversationRepo)
	friendService := service.NewFriendService(userRepo, friendRepo, conversationRepo, idGenerator, nil)
	messageService := service.NewMessageService(conversationRepo, friendRepo, messageRepo, idGenerator, cfg.IM.TextMessageMaxLength)
	onlineService := service.NewOnlineService(onlineRepo, cfg.WebSocket.ServerID, time.Duration(cfg.WebSocket.OnlineTTLSeconds)*time.Second)
	authMiddleware := middleware.NewAuthMiddleware(tokenManager, tokenRepo)
	wsHub := ws.NewHub(onlineService)
	go wsHub.Run(context.Background())
	messageDispatcher := ws.NewSyncMessageDispatcher(messageService, wsHub, time.Duration(cfg.WebSocket.WriteWaitSeconds)*time.Second)
	wsServer := ws.NewServer(wsHub, tokenManager, tokenRepo, messageDispatcher, ws.OptionsFromConfig(cfg.WebSocket))

	engine := router.New(router.Dependencies{
		AuthHandler:         handler.NewAuthHandler(authService),
		UserHandler:         handler.NewUserHandler(userService),
		FriendHandler:       handler.NewFriendHandler(friendService),
		ConversationHandler: handler.NewConversationHandler(conversationService),
		MessageHandler:      handler.NewMessageHandler(messageService),
		AuthMiddleware:      authMiddleware,
		WSServer:            wsServer,
	})
	if err := engine.Run(cfg.Server.Address()); err != nil {
		logger.L().Fatal("server stopped", zap.Error(err))
	}
}
