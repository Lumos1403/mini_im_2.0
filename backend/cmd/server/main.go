package main

import (
	"context"
	"time"

	"mini_im/backend/internal/api/handler"
	"mini_im/backend/internal/api/router"
	"mini_im/backend/internal/config"
	"mini_im/backend/internal/middleware"
	"mini_im/backend/internal/pkg/agentclient"
	jwtpkg "mini_im/backend/internal/pkg/jwt"
	"mini_im/backend/internal/pkg/logger"
	"mini_im/backend/internal/pkg/snowflake"
	mysqlrepo "mini_im/backend/internal/repository/mysql"
	redisrepo "mini_im/backend/internal/repository/redis"
	"mini_im/backend/internal/service"
	"mini_im/backend/internal/storage"
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
	fileRepo := mysqlrepo.NewFileRepository(db)
	groupRepo := mysqlrepo.NewGroupRepository(db)
	tokenRepo := redisrepo.NewTokenRepository(redisClient)
	onlineRepo := redisrepo.NewOnlineRepository(redisClient)
	messageCacheRepo := redisrepo.NewMessageRepository(redisClient)
	fileStorage, err := storage.NewLocalFileStorage(cfg.File.LocalPath)
	if err != nil {
		logger.L().Fatal("file storage init failed", zap.Error(err))
	}
	agentTimeout := time.Duration(cfg.Agent.APITimeoutSeconds) * time.Second
	agentClient := agentclient.New(cfg.Agent.APIBaseURL, agentTimeout)
	agentService := service.NewAgentService(userRepo, friendRepo, conversationRepo, messageRepo, idGenerator, service.AgentOptions{
		Enabled:          cfg.Agent.Enabled,
		APIBaseURL:       cfg.Agent.APIBaseURL,
		APITimeout:       agentTimeout,
		DefaultUsername:  cfg.Agent.DefaultUsername,
		DefaultNickname:  cfg.Agent.DefaultNickname,
		DefaultAvatarURL: cfg.Agent.DefaultAvatarURL,
	}, agentClient)
	agentEnsureCtx, agentEnsureCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defaultAgent, err := agentService.EnsureDefaultAgentUser(agentEnsureCtx)
	agentEnsureCancel()
	if err != nil {
		logger.L().Fatal("default agent user ensure failed", zap.Error(err))
	}
	logger.L().Info("default agent user ensured",
		zap.Int64("agent_user_id", defaultAgent.User.UserID),
		zap.Bool("agent_enabled", cfg.Agent.Enabled),
	)
	authService := service.NewAuthService(userRepo, tokenRepo, tokenManager, idGenerator, agentService)
	userService := service.NewUserService(userRepo)
	conversationService := service.NewConversationService(conversationRepo)
	friendService := service.NewFriendService(userRepo, friendRepo, conversationRepo, idGenerator)
	groupService := service.NewGroupService(groupRepo, idGenerator, cfg.IM.GroupMaxMembers)
	fileService := service.NewFileService(fileRepo, fileStorage, idGenerator, cfg.File.MaxSizeMB)
	messageService := service.NewMessageService(conversationRepo, friendRepo, messageRepo, fileRepo, groupRepo, messageCacheRepo, idGenerator, cfg.IM.TextMessageMaxLength, cfg.IM.RecallMinutes)
	messageService.SetAgentService(agentService)
	searchRepo := mysqlrepo.NewSearchRepository(db)
	searchService := service.NewSearchService(searchRepo, userRepo)
	onlineService := service.NewOnlineService(onlineRepo, cfg.WebSocket.ServerID, time.Duration(cfg.WebSocket.OnlineTTLSeconds)*time.Second)
	authMiddleware := middleware.NewAuthMiddleware(tokenManager, tokenRepo)
	wsHub := ws.NewHub(onlineService)
	go wsHub.Run(context.Background())
	messageService.SetRecallNotifier(ws.NewRecallNotifier(wsHub, time.Duration(cfg.WebSocket.WriteWaitSeconds)*time.Second))
	agentService.SetMessageNotifier(ws.NewAgentNotifier(wsHub, time.Duration(cfg.WebSocket.WriteWaitSeconds)*time.Second))
	messageDispatcher := ws.NewSyncMessageDispatcher(messageService, wsHub, time.Duration(cfg.WebSocket.WriteWaitSeconds)*time.Second)
	wsServer := ws.NewServer(wsHub, tokenManager, tokenRepo, messageDispatcher, ws.OptionsFromConfig(cfg.WebSocket))

	engine := router.New(router.Dependencies{
		AuthHandler:         handler.NewAuthHandler(authService),
		UserHandler:         handler.NewUserHandler(userService),
		FriendHandler:       handler.NewFriendHandler(friendService),
		GroupHandler:        handler.NewGroupHandler(groupService),
		ConversationHandler: handler.NewConversationHandler(conversationService),
		MessageHandler:      handler.NewMessageHandler(messageService),
		FileHandler:         handler.NewFileHandler(fileService),
		SearchHandler:       handler.NewSearchHandler(searchService),
		AuthMiddleware:      authMiddleware,
		WSServer:            wsServer,
	})
	if err := engine.Run(cfg.Server.Address()); err != nil {
		logger.L().Fatal("server stopped", zap.Error(err))
	}
}
