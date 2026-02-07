package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	Port string
	Host string

	// Database
	DatabasePath  string
	WASessionPath string

	// Storage
	TempDir       string
	MaxUploadSize int64
	ChunkSize     int64

	// File settings
	DefaultExpiryDays int
	MaxExpiryDays     int
	ShortIDLength     int

	// Logging
	LogLevel          string
	LogFormat         string
	LogOutput         string
	LogFilePath       string
	LogFileMaxSize    int
	LogFileMaxBackups int

	// Stats
	StatsFlushInterval   time.Duration
	StatsHourlyRetention time.Duration

	// Cleanup jobs
	CleanupInterval     time.Duration
	IncompleteUploadTTL time.Duration

	// Graceful shutdown
	ShutdownTimeout time.Duration

	// Admin auth
	AdminPassword      string
	AdminSessionSecret string
	AdminSessionMaxAge int
}

func Load() *Config {
	return &Config{
		// Server
		Port: getEnv("PORT", "3000"),
		Host: getEnv("HOST", "0.0.0.0"),

		// Database
		DatabasePath:  getEnv("DATABASE_PATH", "./data/whatsbox.db"),
		WASessionPath: getEnv("WA_SESSION_PATH", "./data/wa_session.db"),

		// Storage
		TempDir:       getEnv("TEMP_DIR", "./data/temp"),
		MaxUploadSize: getEnvInt64("MAX_UPLOAD_SIZE", 2147483648), // 2GB
		ChunkSize:     getEnvInt64("CHUNK_SIZE", 10485760),        // 10MB

		// File settings
		DefaultExpiryDays: getEnvInt("DEFAULT_EXPIRY_DAYS", 30),
		MaxExpiryDays:     getEnvInt("MAX_EXPIRY_DAYS", 30),
		ShortIDLength:     getEnvInt("SHORT_ID_LENGTH", 6),

		// Logging
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		LogFormat:         getEnv("LOG_FORMAT", "json"),
		LogOutput:         getEnv("LOG_OUTPUT", "stdout"),
		LogFilePath:       getEnv("LOG_FILE_PATH", "./data/logs/whatsbox.log"),
		LogFileMaxSize:    getEnvInt("LOG_FILE_MAX_SIZE", 100),
		LogFileMaxBackups: getEnvInt("LOG_FILE_MAX_BACKUPS", 10),

		// Stats
		StatsFlushInterval:   time.Duration(getEnvInt("STATS_FLUSH_INTERVAL", 60)) * time.Second,
		StatsHourlyRetention: time.Duration(getEnvInt("STATS_HOURLY_RETENTION", 168)) * time.Hour,

		// Cleanup jobs
		CleanupInterval:     time.Duration(getEnvInt("CLEANUP_INTERVAL", 3600)) * time.Second,
		IncompleteUploadTTL: time.Duration(getEnvInt("INCOMPLETE_UPLOAD_TTL", 86400)) * time.Second,

		// Graceful shutdown
		ShutdownTimeout: time.Duration(getEnvInt("SHUTDOWN_TIMEOUT", 300)) * time.Second,

		// Admin auth
		AdminPassword:      getEnv("ADMIN_PASSWORD", ""),
		AdminSessionSecret: getEnv("ADMIN_SESSION_SECRET", generateDefaultSecret()),
		AdminSessionMaxAge: getEnvInt("ADMIN_SESSION_MAX_AGE", 86400), // 24 hours
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func generateDefaultSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fail hard on random generation failure - this is a security-critical operation
		panic(fmt.Sprintf("Failed to generate secure session secret: %v", err))
	}
	return hex.EncodeToString(bytes)
}
