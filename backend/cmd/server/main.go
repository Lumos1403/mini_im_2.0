package main

import (
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
	tokenRepo := redisrepo.NewTokenRepository(redisClient)
	agentService := service.NewAgentService()
	authService := service.NewAuthService(userRepo, tokenRepo, tokenManager, idGenerator, agentService)
	userService := service.NewUserService(userRepo)
	authMiddleware := middleware.NewAuthMiddleware(tokenManager, tokenRepo)

	engine := router.New(router.Dependencies{
		AuthHandler:    handler.NewAuthHandler(authService),
		UserHandler:    handler.NewUserHandler(userService),
		AuthMiddleware: authMiddleware,
	})
	if err := engine.Run(cfg.Server.Address()); err != nil {
		logger.L().Fatal("server stopped", zap.Error(err))
	}
}
