package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server    ServerConfig
	MySQL     MySQLConfig
	Redis     RedisConfig
	JWT       JWTConfig
	File      FileConfig
	IM        IMConfig
	Snowflake SnowflakeConfig
	WebSocket WebSocketConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

func (c ServerConfig) Address() string {
	return ":" + c.Port
}

type MySQLConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	AccessSecret        string
	RefreshSecret       string
	AccessExpireMinutes int
	RefreshExpireDays   int
}

type FileConfig struct {
	StorageType string
	LocalPath   string
	MaxSizeMB   int
}

type IMConfig struct {
	GroupMaxMembers int
	RecallMinutes   int
}

type SnowflakeConfig struct {
	NodeID int64
}

type WebSocketConfig struct {
	ServerID              string
	WriteWaitSeconds      int
	PongWaitSeconds       int
	PingPeriodSeconds     int
	OnlineTTLSeconds      int
	MaxMessageBytes       int64
	SendBufferSize        int
	AllowedOrigins        []string
	AllowLocalhostOrigins bool
}

func Load() *Config {
	serverMode := getEnv("SERVER_MODE", getEnv("GIN_MODE", ginDebugMode))
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8081"),
			Mode: serverMode,
		},
		MySQL: MySQLConfig{
			DSN: os.Getenv("MYSQL_DSN"),
		},
		Redis: RedisConfig{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			AccessSecret:        os.Getenv("JWT_ACCESS_SECRET"),
			RefreshSecret:       os.Getenv("JWT_REFRESH_SECRET"),
			AccessExpireMinutes: getEnvInt("JWT_ACCESS_EXPIRE_MINUTES", 15),
			RefreshExpireDays:   getEnvInt("JWT_REFRESH_EXPIRE_DAYS", 7),
		},
		File: FileConfig{
			StorageType: getEnv("FILE_STORAGE_TYPE", "local"),
			LocalPath:   getEnv("FILE_LOCAL_PATH", "./uploads"),
			MaxSizeMB:   getEnvInt("FILE_MAX_SIZE_MB", 50),
		},
		IM: IMConfig{
			GroupMaxMembers: getEnvInt("IM_GROUP_MAX_MEMBERS", 50),
			RecallMinutes:   getEnvInt("IM_RECALL_MINUTES", 5),
		},
		Snowflake: SnowflakeConfig{
			NodeID: int64(getEnvInt("SNOWFLAKE_NODE_ID", 1)),
		},
		WebSocket: WebSocketConfig{
			ServerID:              getEnv("WS_SERVER_ID", "ws-1"),
			WriteWaitSeconds:      getEnvInt("WS_WRITE_WAIT_SECONDS", 10),
			PongWaitSeconds:       getEnvInt("WS_PONG_WAIT_SECONDS", 60),
			PingPeriodSeconds:     getEnvInt("WS_PING_PERIOD_SECONDS", 30),
			OnlineTTLSeconds:      getEnvInt("WS_ONLINE_TTL_SECONDS", 60),
			MaxMessageBytes:       int64(getEnvInt("WS_MAX_MESSAGE_BYTES", 65536)),
			SendBufferSize:        getEnvInt("WS_SEND_BUFFER_SIZE", 256),
			AllowedOrigins:        getEnvList("WS_ALLOWED_ORIGINS"),
			AllowLocalhostOrigins: getEnvBool("WS_ALLOW_LOCAL_ORIGINS", serverMode != "release"),
		},
	}
}

const ginDebugMode = "debug"

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvList(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}
