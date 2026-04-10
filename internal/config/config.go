// Package config provides centralized configuration for ADR Insight.
// All settings are loaded from environment variables with sensible defaults.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Config holds all application settings.
type Config struct {
	Port                 int
	DBPath               string
	ADRDir               string
	OllamaURL            string
	AnthropicKey         string
	EmbedModel           string
	LogLevel             slog.Level
	LogFormat            string
	SlowRequestThreshold time.Duration
	ShutdownTimeout      time.Duration
	RateLimitRequests    int
	RateLimitWindow      time.Duration
	MaxQueryLength       int
}

// Load reads configuration from environment variables with defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Port:                 8081,
		DBPath:               "./adr-insight.db",
		ADRDir:               "./docs/adr",
		OllamaURL:            "http://localhost:11434",
		AnthropicKey:         os.Getenv("ANTHROPIC_API_KEY"),
		EmbedModel:           "nomic-embed-text",
		LogLevel:             slog.LevelInfo,
		LogFormat:            "json",
		SlowRequestThreshold: 2 * time.Second,
		ShutdownTimeout:      10 * time.Second,
		RateLimitRequests:    20,
		RateLimitWindow:      time.Hour,
		MaxQueryLength:       200,
	}

	if v := os.Getenv("PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT %q: %w", v, err)
		}
		cfg.Port = p
	}

	if v := os.Getenv("DB_PATH"); v != "" {
		cfg.DBPath = v
	}

	if v := os.Getenv("ADR_DIR"); v != "" {
		cfg.ADRDir = v
	}

	if v := os.Getenv("OLLAMA_URL"); v != "" {
		cfg.OllamaURL = v
	}

	if v := os.Getenv("EMBED_MODEL"); v != "" {
		cfg.EmbedModel = v
	}

	if v := os.Getenv("LOG_FORMAT"); v != "" {
		switch v {
		case "json", "text":
			cfg.LogFormat = v
		default:
			return nil, fmt.Errorf("invalid LOG_FORMAT %q: must be \"json\" or \"text\"", v)
		}
	}

	if v := os.Getenv("LOG_LEVEL"); v != "" {
		switch v {
		case "debug":
			cfg.LogLevel = slog.LevelDebug
		case "info":
			cfg.LogLevel = slog.LevelInfo
		case "warn":
			cfg.LogLevel = slog.LevelWarn
		case "error":
			cfg.LogLevel = slog.LevelError
		default:
			return nil, fmt.Errorf("invalid LOG_LEVEL %q: must be debug, info, warn, or error", v)
		}
	}

	if v := os.Getenv("SLOW_REQUEST_MS"); v != "" {
		ms, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid SLOW_REQUEST_MS %q: %w", v, err)
		}
		cfg.SlowRequestThreshold = time.Duration(ms) * time.Millisecond
	}

	if v := os.Getenv("SHUTDOWN_TIMEOUT_S"); v != "" {
		s, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid SHUTDOWN_TIMEOUT_S %q: %w", v, err)
		}
		cfg.ShutdownTimeout = time.Duration(s) * time.Second
	}

	if v := os.Getenv("RATE_LIMIT_REQUESTS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid RATE_LIMIT_REQUESTS %q: %w", v, err)
		}
		cfg.RateLimitRequests = n
	}

	if v := os.Getenv("RATE_LIMIT_WINDOW_S"); v != "" {
		s, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid RATE_LIMIT_WINDOW_S %q: %w", v, err)
		}
		cfg.RateLimitWindow = time.Duration(s) * time.Second
	}

	if v := os.Getenv("MAX_QUERY_LENGTH"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid MAX_QUERY_LENGTH %q: %w", v, err)
		}
		cfg.MaxQueryLength = n
	}

	return cfg, nil
}

// SetupLogger creates and sets the default slog logger based on config.
func SetupLogger(cfg *Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	var handler slog.Handler
	if cfg.LogFormat == "text" {
		handler = slog.NewTextHandler(os.Stderr, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
