package config

import (
	"os"
	"testing"
)

func TestParseDatabaseURL_Normal(t *testing.T) {
	cfg, err := parseDatabaseURL("postgres://user:pass@localhost:5432/mydb?sslmode=disable")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.User != "user" {
		t.Errorf("User = %q, want %q", cfg.User, "user")
	}
	if cfg.Password != "pass" {
		t.Errorf("Password = %q, want %q", cfg.Password, "pass")
	}
	if cfg.Host != "localhost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "localhost")
	}
	if cfg.Port != "5432" {
		t.Errorf("Port = %q, want %q", cfg.Port, "5432")
	}
	if cfg.Name != "mydb" {
		t.Errorf("Name = %q, want %q", cfg.Name, "mydb")
	}
	if cfg.SSLMode != "disable" {
		t.Errorf("SSLMode = %q, want %q", cfg.SSLMode, "disable")
	}
}

func TestParseDatabaseURL_NoPassword(t *testing.T) {
	cfg, err := parseDatabaseURL("postgres://user@localhost:5432/mydb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.User != "user" {
		t.Errorf("User = %q, want %q", cfg.User, "user")
	}
	if cfg.Password != "" {
		t.Errorf("Password = %q, want empty", cfg.Password)
	}
	if cfg.Name != "mydb" {
		t.Errorf("Name = %q, want %q", cfg.Name, "mydb")
	}
}

func TestParseDatabaseURL_DefaultPort(t *testing.T) {
	cfg, err := parseDatabaseURL("postgres://user@localhost/mydb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "5432" {
		t.Errorf("Port = %q, want %q", cfg.Port, "5432")
	}
}

func TestParseDatabaseURL_InvalidScheme(t *testing.T) {
	_, err := parseDatabaseURL("mysql://user:pass@localhost:3306/mydb")
	if err == nil {
		t.Error("expected error for unsupported scheme, got nil")
	}
}

func TestParseDatabaseURL_NoDatabaseName(t *testing.T) {
	_, err := parseDatabaseURL("postgres://user:pass@localhost:5432/")
	if err == nil {
		t.Error("expected error for missing database name, got nil")
	}
}

func TestParseDatabaseURL_InvalidURL(t *testing.T) {
	_, err := parseDatabaseURL("://invalid")
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}

func TestDBConfig_ToURL_WithPassword(t *testing.T) {
	cfg := dbConfig{
		Host: "localhost", Port: "5432",
		User: "postgres", Password: "secret",
		Name: "testdb", SSLMode: "disable",
	}
	url := cfg.toURL()
	expected := "postgres://postgres:secret@localhost:5432/testdb?sslmode=disable"
	if url != expected {
		t.Errorf("toURL() = %q, want %q", url, expected)
	}
}

func TestDBConfig_ToURL_NoPassword(t *testing.T) {
	cfg := dbConfig{
		Host: "localhost", Port: "5432",
		User: "postgres", Password: "",
		Name: "testdb", SSLMode: "disable",
	}
	url := cfg.toURL()
	expected := "postgres://postgres@localhost:5432/testdb?sslmode=disable"
	if url != expected {
		t.Errorf("toURL() = %q, want %q", url, expected)
	}
}

func TestGetEnv_Default(t *testing.T) {
	os.Unsetenv("TEST_GETENV_KEY")
	val := getEnv("TEST_GETENV_KEY", "fallback")
	if val != "fallback" {
		t.Errorf("getEnv() = %q, want %q", val, "fallback")
	}
}

func TestGetEnv_Override(t *testing.T) {
	os.Setenv("TEST_GETENV_KEY", "custom")
	defer os.Unsetenv("TEST_GETENV_KEY")
	val := getEnv("TEST_GETENV_KEY", "fallback")
	if val != "custom" {
		t.Errorf("getEnv() = %q, want %q", val, "custom")
	}
}

func TestLoadDatabaseURL_FromEnv(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://u:p@h:5432/db")
	defer os.Unsetenv("DATABASE_URL")
	url := loadDatabaseURL()
	if url != "postgres://u:p@h:5432/db" {
		t.Errorf("loadDatabaseURL() = %q, want %q", url, "postgres://u:p@h:5432/db")
	}
}

func TestLoadDatabaseURL_FallbackToDBFields(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Setenv("DB_HOST", "myhost")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_USER", "admin")
	os.Setenv("DB_PASSWORD", "pw")
	os.Setenv("DB_NAME", "mydb")
	os.Setenv("DB_SSLMODE", "require")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
	}()
	url := loadDatabaseURL()
	expected := "postgres://admin:pw@myhost:3306/mydb?sslmode=require"
	if url != expected {
		t.Errorf("loadDatabaseURL() = %q, want %q", url, expected)
	}
}
