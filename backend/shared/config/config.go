package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	// 自动加载项目的 .env 文件，不覆盖已存在的环境变量
	// 这样宝塔里配的 APP_PORT 等会保留，JWT_SECRET 等从文件补充
	_ = godotenv.Load("/www/wwwroot/fettle-standalone/.env")
}

// AppConfig 应用通用配置
type AppConfig struct {
	Port       int    `yaml:"port"`
	Name       string `yaml:"name"`
	Env        string `yaml:"env"`
	JWTSecret  string `yaml:"jwt_secret"`
	LogLevel   string `yaml:"log_level"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// GetEnv 获取环境变量，不存在返回默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt 获取整型环境变量
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvBool 获取布尔环境变量
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
