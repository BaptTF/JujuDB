package config

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

// Config holds all application configuration
type Config struct {
	Database    DatabaseConfig
	Server      ServerConfig
	Meilisearch MeilisearchConfig
	Session     SessionConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	Name     string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port       string
	Production bool
	HTTPS      bool
}

// MeilisearchConfig holds Meilisearch configuration
type MeilisearchConfig struct {
	Host      string
	MasterKey string
}

// SessionConfig holds session configuration
type SessionConfig struct {
	Key    string
	Secure bool
	MaxAge int
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			User:     getEnv("DB_USER", "jujudb"),
			Password: getEnv("DB_PASSWORD", "jujudb123"),
			Name:     getEnv("DB_NAME", "jujudb"),
		},
		Server: ServerConfig{
			Port:       getEnv("PORT", "8080"),
			Production: getEnv("PRODUCTION", "false") == "true",
			HTTPS:      getEnv("HTTPS", "false") == "true",
		},
		Meilisearch: MeilisearchConfig{
			Host:      getEnv("MEILI_HOST", "http://localhost:7700"),
			MasterKey: getEnv("MEILI_MASTER_KEY", "jujudb-master-key-change-in-production"),
		},
		Session: SessionConfig{
			Key:    getSessionKey(),
			Secure: getEnv("PRODUCTION", "false") == "true" || getEnv("HTTPS", "false") == "true",
			MaxAge: 86400 * 30, // 30 days
		},
	}

	return config
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getSessionKey gets session key with production warning
func getSessionKey() string {
	defaultKey := "dev-session-key-change-in-production"

	if key := os.Getenv("SESSION_KEY"); key != "" {
		return key
	}

	// Warning for production
	if os.Getenv("PRODUCTION") == "true" {
		logrus.Warn("WARNING: Using default session key in production! Please set SESSION_KEY environment variable.")
	}

	return defaultKey
}

// GetConnectionString returns database connection string
func (d *DatabaseConfig) GetConnectionString() string {
	return "host=" + d.Host + " user=" + d.User + " password=" + d.Password + " dbname=" + d.Name + " sslmode=disable"
}

// GetPortAsInt returns port as integer
func (s *ServerConfig) GetPortAsInt() int {
	if port, err := strconv.Atoi(s.Port); err == nil {
		return port
	}
	return 8080
}
