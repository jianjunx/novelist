package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    DatabaseURL   string
    JWTSecret     string
    ServerPort    string
    DeepSeekKey   string
    DeepSeekModel string
}

func Load() *Config {
    // 加载.env文件（如果存在）
    _ = godotenv.Load()

    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        log.Fatal("JWT_SECRET environment variable is required")
    }

    return &Config{
        DatabaseURL:   getEnv("DATABASE_URL", "postgres://localhost:5432/novelist?sslmode=disable"),
        JWTSecret:     jwtSecret,
        ServerPort:    getEnv("SERVER_PORT", "8080"),
        DeepSeekKey:   os.Getenv("DEEPSEEK_API_KEY"),
        DeepSeekModel: getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
