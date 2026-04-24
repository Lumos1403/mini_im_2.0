package ws

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mini_im/backend/internal/config"
	apperrors "mini_im/backend/internal/pkg/errors"
	jwtpkg "mini_im/backend/internal/pkg/jwt"
	"mini_im/backend/internal/pkg/logger"
	"mini_im/backend/internal/pkg/response"
	redisrepo "mini_im/backend/internal/repository/redis"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Options struct {
	WriteWait             time.Duration
	PongWait              time.Duration
	PingPeriod            time.Duration
	MaxMessageBytes       int64
	SendBufferSize        int
	AllowedOrigins        []string
	AllowLocalhostOrigins bool
}

type Server struct {
	hub          *Hub
	tokenManager *jwtpkg.Manager
	tokenRepo    *redisrepo.TokenRepository
	upgrader     websocket.Upgrader
	options      Options
}

func NewServer(hub *Hub, tokenManager *jwtpkg.Manager, tokenRepo *redisrepo.TokenRepository, options Options) *Server {
	options = normalizeOptions(options)
	return &Server{
		hub:          hub,
		tokenManager: tokenManager,
		tokenRepo:    tokenRepo,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     checkOrigin(options),
		},
		options: options,
	}
}

func OptionsFromConfig(cfg config.WebSocketConfig) Options {
	return Options{
		WriteWait:             time.Duration(cfg.WriteWaitSeconds) * time.Second,
		PongWait:              time.Duration(cfg.PongWaitSeconds) * time.Second,
		PingPeriod:            time.Duration(cfg.PingPeriodSeconds) * time.Second,
		MaxMessageBytes:       cfg.MaxMessageBytes,
		SendBufferSize:        cfg.SendBufferSize,
		AllowedOrigins:        cfg.AllowedOrigins,
		AllowLocalhostOrigins: cfg.AllowLocalhostOrigins,
	}
}

func (s *Server) Handle(ctx *gin.Context) {
	tokenValue := strings.TrimSpace(ctx.Query("token"))
	if tokenValue == "" {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	claims, err := s.tokenManager.ParseAccessToken(tokenValue)
	if err != nil {
		appErr := apperrors.ErrTokenInvalid
		if errors.Is(err, jwtpkg.ErrExpiredToken) {
			appErr = apperrors.ErrTokenExpired
		}
		response.Fail(ctx, apperrors.HTTPStatus(appErr), appErr)
		return
	}

	blacklisted, err := s.tokenRepo.IsAccessTokenBlacklisted(ctx.Request.Context(), claims.JTI)
	if err != nil {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrInternal), apperrors.ErrInternal)
		return
	}
	if blacklisted {
		response.Fail(ctx, apperrors.HTTPStatus(apperrors.ErrTokenInvalid), apperrors.ErrTokenInvalid)
		return
	}

	conn, err := s.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.L().Warn("websocket upgrade failed", zap.Error(err))
		return
	}

	client := NewClient(claims.UserID, conn, s.hub, s.options)
	registerCtx, cancel := context.WithTimeout(ctx.Request.Context(), s.options.WriteWait)
	defer cancel()
	if err := s.hub.Register(registerCtx, client); err != nil {
		logger.L().Warn("websocket register failed", zap.Int64("user_id", claims.UserID), zap.Error(err))
		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "register failed"),
			time.Now().Add(s.options.WriteWait),
		)
		_ = conn.Close()
		return
	}

	logger.L().Info("websocket connected", zap.Int64("user_id", claims.UserID), zap.String("remote_addr", ctx.ClientIP()))
	go client.WritePump()
	go client.ReadPump()
}

func normalizeOptions(options Options) Options {
	if options.WriteWait <= 0 {
		options.WriteWait = 10 * time.Second
	}
	if options.PongWait <= 0 {
		options.PongWait = 60 * time.Second
	}
	if options.PingPeriod <= 0 || options.PingPeriod >= options.PongWait {
		options.PingPeriod = 30 * time.Second
	}
	if options.MaxMessageBytes <= 0 {
		options.MaxMessageBytes = 64 * 1024
	}
	if options.SendBufferSize <= 0 {
		options.SendBufferSize = 256
	}
	return options
}

func checkOrigin(options Options) func(req *http.Request) bool {
	allowedHosts := make(map[string]struct{}, len(options.AllowedOrigins))
	for _, origin := range options.AllowedOrigins {
		host := normalizeOriginHost(origin)
		if host != "" {
			allowedHosts[host] = struct{}{}
		}
	}

	return func(req *http.Request) bool {
		origin := req.Header.Get("Origin")
		if origin == "" {
			return options.AllowLocalhostOrigins
		}

		originHost := normalizeOriginHost(origin)
		if originHost == "" {
			return false
		}
		if _, ok := allowedHosts[originHost]; ok {
			return true
		}

		requestHost := strings.ToLower(hostWithoutPort(req.Host))
		return options.AllowLocalhostOrigins && isLocalhost(originHost) && isLocalhost(requestHost)
	}
}

func normalizeOriginHost(origin string) string {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return ""
	}

	originURL, err := url.Parse(origin)
	if err == nil && originURL.Hostname() != "" {
		return strings.ToLower(originURL.Hostname())
	}

	return strings.ToLower(hostWithoutPort(origin))
}

func hostWithoutPort(host string) string {
	hostname, _, err := net.SplitHostPort(host)
	if err == nil {
		return hostname
	}
	return host
}

func isLocalhost(host string) bool {
	host = strings.ToLower(host)
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}
