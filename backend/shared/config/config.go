package config

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	// 自动加载 .env（不覆盖已有环境变量）：
	// 1) 优先取执行文件所在目录的 .env（支持 fettle / fettle-standalone 各自独立部署）
	// 2) 回退到当前工作目录的 .env
	// 3) 沿父目录向上查找至 /.env（兼容以子目录为根的多服务部署）
	candidates := envCandidates()
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			_ = godotenv.Load(p)
			return
		}
	}
}

func envCandidates() []string {
	candidates := make([]string, 0, 8)
	if exe, err := os.Executable(); err == nil && exe != "" {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), ".env"))
	}
	if cwd, err := os.Getwd(); err == nil && cwd != "" {
		candidates = append(candidates, filepath.Join(cwd, ".env"))
		for _, p := range walkUpToRoot(cwd) {
			candidates = append(candidates, filepath.Join(p, ".env"))
		}
	}
	return candidates
}

// walkUpToRoot 从 start 逐级向上，返回每一级目录（含 start 与盘符/挂载根），用于查找上层 .env
func walkUpToRoot(start string) []string {
	var dirs []string
	cur := start
	for {
		dirs = append(dirs, cur)
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return dirs
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
