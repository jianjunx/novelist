package config

import (
    "fmt"
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

    // 构建数据库连接URL
    dbHost := getEnv("DB_HOST", "localhost")
    dbPort := getEnv("DB_PORT", "5432")
    dbUser := getEnv("DB_USER", "postgres")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbName := getEnv("DB_NAME", "novelist")
    dbSSLMode := getEnv("DB_SSLMODE", "disable")

    var databaseURL string
    if dbPassword != "" {
        databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)
    } else {
        databaseURL = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s", dbUser, dbHost, dbPort, dbName, dbSSLMode)
    }

    return &Config{
        DatabaseURL:   databaseURL,
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
