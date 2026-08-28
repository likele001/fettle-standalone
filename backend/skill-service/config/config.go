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
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
}

func Load() *Config {
	port := sharedconfig.GetEnvInt("APP_PORT", 9500)
	dbPort := sharedconfig.GetEnvInt("DB_PORT", 5432)

	return &Config{
		Port:     port,
		DBHost:   sharedconfig.GetEnv("DB_HOST", "localhost"),
		DBPort:   dbPort,
		DBUser:   sharedconfig.GetEnv("DB_USER", "ai_platform"),
		DBPass:   sharedconfig.GetEnv("DB_PASSWORD", "CHANGE_ME"),
		DBName:   sharedconfig.GetEnv("DB_NAME", "ai_platform"),
		MinIOEndpoint:  sharedconfig.GetEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: sharedconfig.GetEnv("MINIO_ACCESS_KEY", "ai_platform"),
		MinIOSecretKey: sharedconfig.GetEnv("MINIO_SECRET_KEY", "CHANGE_ME"),
	}
}

func (c *Config) DatabaseDSN() string {
	return "host=" + c.DBHost + " port=" + strconv.Itoa(c.DBPort) +
		" user=" + c.DBUser + " password=" + c.DBPass +
		" dbname=" + c.DBName + " sslmode=disable"
}
