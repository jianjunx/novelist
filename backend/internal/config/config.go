package config

import "os"

type Config struct {
    DatabaseURL   string
    JWTSecret     string
    ServerPort    string
    DeepSeekKey   string
    DeepSeekModel string
}

func Load() *Config {
    return &Config{
        DatabaseURL:   getEnv("DATABASE_URL", "postgres://localhost:5432/novelist?sslmode=disable"),
        JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
        ServerPort:    getEnv("SERVER_PORT", "8080"),
        DeepSeekKey:   getEnv("DEEPSEEK_API_KEY", ""),
        DeepSeekModel: getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
