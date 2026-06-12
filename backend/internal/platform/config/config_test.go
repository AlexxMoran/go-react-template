package config

import "testing"

func TestLoadReadsOptionalTemplateFlags(t *testing.T) {
	t.Setenv("JWT_ACCESS_SECRET", "test-access-secret-test-access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-test-refresh-secret")
	t.Setenv("METRICS_ENABLED", "true")
	t.Setenv("RATE_LIMIT_ENABLED", "false")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com, https://admin.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if !cfg.Metrics.Enabled {
		t.Fatal("metrics should be enabled")
	}
	if cfg.RateLimit.Enabled {
		t.Fatal("rate limit should be disabled")
	}
	if len(cfg.CORS.AllowedOrigins) != 2 || cfg.CORS.AllowedOrigins[1] != "https://admin.example.com" {
		t.Fatalf("allowed origins = %#v, want two trimmed origins", cfg.CORS.AllowedOrigins)
	}
}

func TestLoadRejectsInsecureProductionCookie(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_ACCESS_SECRET", "test-access-secret-test-access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-test-refresh-secret")
	t.Setenv("COOKIE_SECURE", "false")

	if _, err := Load(); err == nil {
		t.Fatal("expected production config with insecure cookies to fail")
	}
}

func TestLoadRejectsShortProductionSecrets(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_ACCESS_SECRET", "short")
	t.Setenv("JWT_REFRESH_SECRET", "short")
	t.Setenv("COOKIE_SECURE", "true")

	if _, err := Load(); err == nil {
		t.Fatal("expected production config with short JWT secrets to fail")
	}
}
