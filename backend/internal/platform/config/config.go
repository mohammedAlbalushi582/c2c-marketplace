// Package config loads application configuration from environment variables
// (optionally sourced from a .env file in development).
package config

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds all runtime configuration. Fields are populated from env vars.
type Config struct {
	AppEnv   string `envconfig:"APP_ENV" default:"development"`
	HTTPPort string `envconfig:"HTTP_PORT" default:"8080"`

	// DatabaseURL is a libpq/pgx DSN, e.g.
	// postgres://user:pass@localhost:5432/alamjad?sslmode=disable
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	// JWT settings (used from Deliverable #2 onward).
	JWTAccessSecret  string        `envconfig:"JWT_ACCESS_SECRET" default:"dev-access-secret-change-me"`
	JWTRefreshSecret string        `envconfig:"JWT_REFRESH_SECRET" default:"dev-refresh-secret-change-me"`
	JWTAccessTTL     time.Duration `envconfig:"JWT_ACCESS_TTL" default:"15m"`
	JWTRefreshTTL    time.Duration `envconfig:"JWT_REFRESH_TTL" default:"720h"`

	// CORS allowed origins (comma-separated), e.g. the Next.js dev origin.
	CORSAllowedOrigins []string `envconfig:"CORS_ALLOWED_ORIGINS" default:"http://localhost:3000"`

	// Local file-upload storage (swappable for S3/Oracle later).
	UploadDir     string `envconfig:"UPLOAD_DIR" default:"./uploads"`
	PublicBaseURL string `envconfig:"PUBLIC_BASE_URL" default:"http://localhost:8080"`

	// Frontend origin — used to build payment success/cancel redirect URLs.
	FrontendBaseURL string `envconfig:"FRONTEND_BASE_URL" default:"http://localhost:3000"`

	// Payment gateway. "stub" (default) simulates payments for dev/demo; "thawani"
	// enables the real Thawani Pay checkout once the keys below are set.
	PaymentProvider       string `envconfig:"PAYMENT_PROVIDER" default:"stub"`
	ThawaniBaseURL        string `envconfig:"THAWANI_BASE_URL" default:"https://uatcheckout.thawani.om/api/v1"`
	ThawaniSecretKey      string `envconfig:"THAWANI_SECRET_KEY" default:""`
	ThawaniPublishableKey string `envconfig:"THAWANI_PUBLISHABLE_KEY" default:""`
}

// Load reads a .env file when present (dev convenience) then parses env vars.
// It searches the current dir and the repo root (one level up) so it works
// whether the API is run from ./backend or the monorepo root. A missing .env
// is not an error; real env vars always win (e.g. in Docker/production).
func Load() (*Config, error) {
	for _, p := range []string{".env", "../.env"} {
		if err := godotenv.Load(p); err == nil {
			break
		}
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// IsProduction reports whether the app is running in a production environment.
func (c *Config) IsProduction() bool { return c.AppEnv == "production" }
