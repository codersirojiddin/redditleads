package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	AppEnv string
	Port   string

	DatabaseURL string

	RedditClientID     string
	RedditClientSecret string
	RedditUserAgent    string

	OpenAIAPIKey string
	OpenAIModel  string

	JWTSecret    string
	JWTExpiryHrs int

	FrontendURL string

	// WorkerPollIntervalSec controls how often the worker polls for pending reports.
	WorkerPollIntervalSec int
	// MaxPostsPerKeyword caps how many Reddit posts we fetch per keyword per run.
	MaxPostsPerKeyword int
	// RunWorkerInProcess makes `apps/api` also run the report-processing
	// worker loop as a background goroutine, so a single deployed service
	// handles both HTTP and background jobs (useful on free hosting tiers
	// that only grant one service). Defaults to true; set to false to run
	// `apps/worker` as a fully separate process instead.
	RunWorkerInProcess bool
}

// Load reads a .env file if present (ignored if missing) and builds a Config
// from environment variables, applying sane defaults and validating required
// values.
func Load() (*Config, error) {
	_ = godotenv.Load() // ok if .env doesn't exist (e.g. in prod containers)

	cfg := &Config{
		AppEnv: getEnv("APP_ENV", "development"),
		Port:   getEnv("PORT", "8080"),

		DatabaseURL: os.Getenv("DATABASE_URL"),

		RedditClientID:     os.Getenv("REDDIT_CLIENT_ID"),
		RedditClientSecret: os.Getenv("REDDIT_CLIENT_SECRET"),
		RedditUserAgent:    getEnv("REDDIT_USER_AGENT", "reddit-leads/0.1 (client)"),

		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:  getEnv("OPENAI_MODEL", "gpt-4o-mini"),

		JWTSecret:    getEnv("JWT_SECRET", ""),
		JWTExpiryHrs: getEnvInt("JWT_EXPIRY_HOURS", 24*7),

		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		WorkerPollIntervalSec: getEnvInt("WORKER_POLL_INTERVAL_SECONDS", 5),
		MaxPostsPerKeyword:    getEnvInt("MAX_POSTS_PER_KEYWORD", 50),
		RunWorkerInProcess:    getEnvBool("RUN_WORKER_INMEMORY", true),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
