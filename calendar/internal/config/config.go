package config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	ArchiveInterval time.Duration
	LogBuffer       int
}

func MustLoad() Config {
	portDefault := env("PORT", "8080")
	archiveMinutesDefault := envInt("ARCHIVE_INTERVAL_MINUTES", 1)
	logBufferDefault := envInt("LOG_BUFFER", 128)

	port := flag.String("port", portDefault, "HTTP server port")
	archiveMinutes := flag.Int("archive-interval-minutes", archiveMinutesDefault, "archive worker interval in minutes")
	logBuffer := flag.Int("log-buffer", logBufferDefault, "async logger buffer size")
	flag.Parse()

	return Config{
		Port:            *port,
		ArchiveInterval: time.Duration(*archiveMinutes) * time.Minute,
		LogBuffer:       *logBuffer,
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
