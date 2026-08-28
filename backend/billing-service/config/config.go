package config

import (
	"strconv"

	sharedconfig "ai-platform/shared/config"
)

type Config struct {
	Port     int
	DBHost   string
	DBPort   int
	DBUser   string
	DBPass   string
	DBName   string
	RedisAddr string
}

func Load() *Config {
	port := sharedconfig.GetEnvInt("APP_PORT", 9600)
	dbPort := sharedconfig.GetEnvInt("DB_PORT", 5432)

	return &Config{
		Port:     port,
		DBHost:   sharedconfig.GetEnv("DB_HOST", "localhost"),
		DBPort:   dbPort,
		DBUser:   sharedconfig.GetEnv("DB_USER", "ai_platform"),
		DBPass:   sharedconfig.GetEnv("DB_PASSWORD", "CHANGE_ME"),
		DBName:   sharedconfig.GetEnv("DB_NAME", "ai_platform"),
		RedisAddr: sharedconfig.GetEnv("REDIS_ADDR", "localhost:6379"),
	}
}

func (c *Config) DatabaseDSN() string {
	return "host=" + c.DBHost + " port=" + strconv.Itoa(c.DBPort) +
		" user=" + c.DBUser + " password=" + c.DBPass +
		" dbname=" + c.DBName + " sslmode=disable"
}
