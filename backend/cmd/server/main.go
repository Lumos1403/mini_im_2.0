package main

import (
	"mini_im/backend/internal/api/router"
	"mini_im/backend/internal/config"
	"mini_im/backend/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	gin.SetMode(cfg.Server.Mode)
	logger.Init(cfg.Server.Mode)
	defer logger.Sync()

	engine := router.New()
	if err := engine.Run(cfg.Server.Address()); err != nil {
		logger.L().Fatal("server stopped", zap.Error(err))
	}
}
