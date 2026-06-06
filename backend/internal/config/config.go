package config

import (
    "fmt"
    "log"
    "net/url"
    "os"
    "strings"

    "github.com/joho/godotenv"
)

type Config struct {
    DatabaseURL      string
    JWTSecret        string
    ServerPort       string
    DeepSeekKey      string
    DeepSeekModel    string
    OpenAIKey        string
    EmbeddingModel   string
    EmbeddingBaseURL string
}

func Load() *Config {
    // 加载.env文件（如果存在）
    _ = godotenv.Load()

    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        log.Fatal("JWT_SECRET environment variable is required")
    }

    databaseURL := loadDatabaseURL()

    return &Config{
        DatabaseURL:      databaseURL,
        JWTSecret:        jwtSecret,
        ServerPort:       getEnv("SERVER_PORT", "8080"),
        DeepSeekKey:      os.Getenv("DEEPSEEK_API_KEY"),
        DeepSeekModel:    getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
        OpenAIKey:        os.Getenv("OPENAI_API_KEY"),
        EmbeddingModel:   getEnv("EMBEDDING_MODEL", "text-embedding-3-small"),
        EmbeddingBaseURL: os.Getenv("EMBEDDING_BASE_URL"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

type dbConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    Name     string
    SSLMode  string
}

func (d dbConfig) toURL() string {
    if d.Password != "" {
        return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
    }
    return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s", d.User, d.Host, d.Port, d.Name, d.SSLMode)
}

func loadDatabaseURL() string {
    if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
        if _, err := parseDatabaseURL(databaseURL); err != nil {
            log.Fatalf("invalid DATABASE_URL: %v", err)
        }
        return databaseURL
    }

    cfg := dbConfig{
        Host:     getEnv("DB_HOST", "localhost"),
        Port:     getEnv("DB_PORT", "5432"),
        User:     getEnv("DB_USER", "postgres"),
        Password: os.Getenv("DB_PASSWORD"),
        Name:     getEnv("DB_NAME", "novelist"),
        SSLMode:  getEnv("DB_SSLMODE", "disable"),
    }
    return cfg.toURL()
}

func parseDatabaseURL(databaseURL string) (*dbConfig, error) {
    u, err := url.Parse(databaseURL)
    if err != nil {
        return nil, err
    }
    if u.Scheme != "postgres" && u.Scheme != "postgresql" {
        return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
    }

    cfg := &dbConfig{
        Host:    u.Hostname(),
        Port:    u.Port(),
        SSLMode: "disable",
    }
    if cfg.Port == "" {
        cfg.Port = "5432"
    }
    if u.User != nil {
        cfg.User = u.User.Username()
        cfg.Password, _ = u.User.Password()
    }
    cfg.Name = strings.TrimPrefix(u.Path, "/")
    if cfg.Name == "" {
        return nil, fmt.Errorf("database name is required")
    }
    if sslmode := u.Query().Get("sslmode"); sslmode != "" {
        cfg.SSLMode = sslmode
    }

    return cfg, nil
}

