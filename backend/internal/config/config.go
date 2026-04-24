package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server    ServerConfig
	MySQL     MySQLConfig
	Redis     RedisConfig
	JWT       JWTConfig
	File      FileConfig
	IM        IMConfig
	Snowflake SnowflakeConfig
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

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8081"),
			Mode: getEnv("SERVER_MODE", getEnv("GIN_MODE", ginDebugMode)),
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
