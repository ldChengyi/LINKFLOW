package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Log      LogConfig
	MQTT     MQTTConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host          string
	Port          string
	Name          string
	AdminUser     string
	AdminPassword string
	AppUser       string
	AppPassword   string
	ReadUser      string
	ReadPassword  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret      string
	ExpireHours int
}

type LogConfig struct {
	Level      string
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

type MQTTConfig struct {
	Host string
	Port string
}

func Load() (*Config, error) {
	// 加载 .env 文件（如果存在）
	_ = godotenv.Load()

	expireHours, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "24"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	logMaxSize, _ := strconv.Atoi(getEnv("LOG_MAX_SIZE", "100"))
	logMaxBackups, _ := strconv.Atoi(getEnv("LOG_MAX_BACKUPS", "3"))
	logMaxAge, _ := strconv.Atoi(getEnv("LOG_MAX_AGE", "7"))
	logCompress := getEnv("LOG_COMPRESS", "true") == "true"

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:          getEnv("DB_HOST", "localhost"),
			Port:          getEnv("DB_PORT", "5432"),
			Name:          getEnv("DB_NAME", "linkflow"),
			AdminUser:     getEnv("DB_ADMIN_USER", "linkflow_admin"),
			AdminPassword: getEnv("DB_ADMIN_PASSWORD", ""),
			AppUser:       getEnv("DB_APP_USER", "linkflow_app"),
			AppPassword:   getEnv("DB_APP_PASSWORD", ""),
			ReadUser:      getEnv("DB_READ_USER", "linkflow_read"),
			ReadPassword:  getEnv("DB_READ_PASSWORD", ""),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", ""),
			ExpireHours: expireHours,
		},
		Log: LogConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Filename:   getEnv("LOG_FILENAME", "logs/linkflow.log"),
			MaxSize:    logMaxSize,
			MaxBackups: logMaxBackups,
			MaxAge:     logMaxAge,
			Compress:   logCompress,
		},
		MQTT: MQTTConfig{
			Host: getEnv("MQTT_HOST", "0.0.0.0"),
			Port: getEnv("MQTT_PORT", "1883"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
