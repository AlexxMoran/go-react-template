// Package config loads all runtime configuration from environment variables.
//
// This mirrors the Python backend's pydantic-settings layer: a single typed
// Config struct, grouped into sub-structs, loaded once at startup. In local
// development a .env file is loaded first (via godotenv) for convenience; in
// production the process environment is the single source of truth.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env       string
	Log       LogConfig
	HTTP      HTTPConfig
	DB        DBConfig
	JWT       JWTConfig
	Cookie    CookieConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
}

type LogConfig struct {
	Level string
}

type HTTPConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	// RequestTimeout bounds how long a single handler may run; it is applied as a
	// context deadline on each request so downstream DB calls are cancelled.
	RequestTimeout time.Duration
	// MaxBodyBytes caps the request body size (415/400 beyond it).
	MaxBodyBytes int64
}

// RateLimitConfig configures the per-client-IP token-bucket rate limiter.
type RateLimitConfig struct {
	Enabled bool
	// RPS is the sustained requests-per-second allowed per client IP.
	RPS float64
	// Burst is the maximum momentary burst above the sustained rate.
	Burst int
}

func (h HTTPConfig) Addr() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	MaxConns int32
	MinConns int32
}

// DSN builds a pgx-compatible PostgreSQL connection string.
func (d DBConfig) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(d.User, d.Password),
		Host:   fmt.Sprintf("%s:%d", d.Host, d.Port),
		Path:   d.Name,
	}
	q := url.Values{}
	q.Set("sslmode", d.SSLMode)
	q.Set("pool_max_conns", strconv.Itoa(int(d.MaxConns)))
	q.Set("pool_min_conns", strconv.Itoa(int(d.MinConns)))
	u.RawQuery = q.Encode()
	return u.String()
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	Issuer        string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type CookieConfig struct {
	Secure   bool
	SameSite string
	Domain   string
}

type CORSConfig struct {
	AllowedOrigins []string
}

func (c Config) IsProduction() bool { return c.Env == "production" }

// Load reads configuration from the environment. It optionally loads a .env
// file first (ignored if absent) and validates required values.
func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Env: getenv("APP_ENV", "development"),
		Log: LogConfig{
			Level: getenv("LOG_LEVEL", "info"),
		},
		HTTP: HTTPConfig{
			Host:            getenv("HTTP_HOST", "0.0.0.0"),
			Port:            getenvInt("HTTP_PORT", 3000),
			ReadTimeout:     getenvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getenvDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: getenvDuration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second),
			RequestTimeout:  getenvDuration("HTTP_REQUEST_TIMEOUT", 15*time.Second),
			MaxBodyBytes:    getenvInt64("HTTP_MAX_BODY_BYTES", 1<<20), // 1 MiB
		},
		DB: DBConfig{
			Host:     getenv("DB_HOST", "localhost"),
			Port:     getenvInt("DB_PORT", 5432),
			User:     getenv("DB_USER", "postgres"),
			Password: getenv("DB_PASSWORD", "postgres"),
			Name:     getenv("DB_NAME", "goapp"),
			SSLMode:  getenv("DB_SSLMODE", "disable"),
			MaxConns: int32(getenvInt("DB_MAX_CONNS", 20)),
			MinConns: int32(getenvInt("DB_MIN_CONNS", 2)),
		},
		JWT: JWTConfig{
			AccessSecret:  os.Getenv("JWT_ACCESS_SECRET"),
			RefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),
			Issuer:        getenv("JWT_ISSUER", "goapp"),
			AccessTTL:     getenvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:    getenvDuration("JWT_REFRESH_TTL", 168*time.Hour),
		},
		Cookie: CookieConfig{
			Secure:   getenvBool("COOKIE_SECURE", false),
			SameSite: getenv("COOKIE_SAMESITE", "lax"),
			Domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getenvCSV("CORS_ALLOWED_ORIGINS"),
		},
		RateLimit: RateLimitConfig{
			Enabled: getenvBool("RATE_LIMIT_ENABLED", true),
			RPS:     getenvFloat("RATE_LIMIT_RPS", 20),
			Burst:   getenvInt("RATE_LIMIT_BURST", 40),
		},
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	if c.JWT.AccessSecret == "" || c.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET are required")
	}
	if c.IsProduction() {
		if len(c.JWT.AccessSecret) < 32 || len(c.JWT.RefreshSecret) < 32 {
			return fmt.Errorf("JWT secrets must be at least 32 bytes in production")
		}
		if !c.Cookie.Secure {
			return fmt.Errorf("COOKIE_SECURE must be true in production")
		}
	}
	return nil
}

// ── env helpers ──────────────────────────────────────────────────────────────

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getenvInt64(key string, fallback int64) int64 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}

func getenvFloat(key string, fallback float64) float64 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getenvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getenvCSV(key string) []string {
	raw := os.Getenv(key)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
