package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Database DatabaseConfig
	Reddit   RedditConfig
	Discord  DiscordConfig
	App      AppConfig
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	URL      string
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	MaxConns int
	MinConns int
}

// RedditConfig holds Reddit API configuration
type RedditConfig struct {
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	UserAgent    string
	Subreddits   []string
}

// DiscordConfig holds Discord bot configuration
type DiscordConfig struct {
	Token                string
	HiringChannelID      string
	TestChannelID        string
	DeveloperChannelID   string
	PublicKey            string
	Enabled              bool
}

// AppConfig holds general application configuration
type AppConfig struct {
	LogLevel        string
	Environment     string
	GracefulTimeout time.Duration
	HackerNewsID    int
	MaxJobsToGet    int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env doesn't exist in production
		if os.Getenv("ENVIRONMENT") != "production" {
			fmt.Printf("Warning: .env file not found: %v\n", err)
		}
	}

	config := &Config{}

	// Database configuration
	config.Database = DatabaseConfig{
		URL:      getEnv("DATABASE_URL", ""),
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("POSTGRES_USER", "postgres"),
		Password: getEnv("POSTGRES_PASSWORD", ""),
		Database: getEnv("POSTGRES_DB", "job_scraper"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		MaxConns: getEnvAsInt("DB_MAX_CONNS", 25),
		MinConns: getEnvAsInt("DB_MIN_CONNS", 5),
	}

	// If DATABASE_URL is not provided, construct it
	if config.Database.URL == "" {
		config.Database.URL = fmt.Sprintf(
			"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
			config.Database.User,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Database,
			config.Database.SSLMode,
		)
	}

	// Reddit configuration
	config.Reddit = RedditConfig{
		ClientID:     getEnv("REDDIT_ID", ""),
		ClientSecret: getEnv("REDDIT_SECRET", ""),
		Username:     getEnv("REDDIT_USERNAME", ""),
		Password:     getEnv("REDDIT_PASSWORD", ""),
		UserAgent:   getEnv("REDDIT_USER_AGENT", "SALPhBot"),
		Subreddits:   []string{"forhire", "remotejs"},
	}

	// Discord configuration
	config.Discord = DiscordConfig{
		Token:              getEnv("JOBBYMCJOBFACE_DISCORD_TOKEN", ""),
		HiringChannelID:    getEnv("JOBBYMCJOBFACE_HIRING_CHANNEL_ID", ""),
		TestChannelID:      getEnv("JOBBYMCJOBFACE_TEST_CHANNEL_ID", ""),
		DeveloperChannelID: getEnv("JOBBYMCJOBFACE_DEVELOPER_CHANNEL_ID", ""),
		PublicKey:          getEnv("JOBBYMCJOBFACE_PUBLIC_KEY", ""),
		Enabled:            getEnvAsBool("DISCORD_ENABLED", true),
	}

	// App configuration
	config.App = AppConfig{
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		GracefulTimeout: getEnvAsDuration("GRACEFUL_TIMEOUT", 30*time.Second),
		HackerNewsID:    getEnvAsInt("HACKERNEWS_CURRENT_ID", 42575537),
		MaxJobsToGet:    getEnvAsInt("MAX_JOBS_TO_GET", 30),
	}

	// Validate required configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate checks that required configuration values are present
func (c *Config) Validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("database URL is required")
	}

	if c.Reddit.ClientID == "" || c.Reddit.ClientSecret == "" {
		return fmt.Errorf("Reddit API credentials are required")
	}

	if c.Discord.Enabled && c.Discord.Token == "" {
		return fmt.Errorf("Discord token is required when Discord is enabled")
	}

	return nil
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}