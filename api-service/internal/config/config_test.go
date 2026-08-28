package config

import (
	"strings"
	"testing"
	"time"
)

var optionalEnvironment = []string{
	"HTTP_ADDRESS",
	"HTTP_READ_HEADER_TIMEOUT",
	"HTTP_READ_TIMEOUT",
	"HTTP_WRITE_TIMEOUT",
	"HTTP_IDLE_TIMEOUT",
	"HTTP_SHUTDOWN_TIMEOUT",
	"DB_MAX_CONNS",
	"DB_MIN_CONNS",
	"DB_MAX_CONN_LIFETIME",
	"DB_MAX_CONN_IDLE_TIME",
	"DB_HEALTH_CHECK_PERIOD",
	"DB_CONNECT_TIMEOUT",
}

func TestLoadDefaults(t *testing.T) {
	clearOptionalEnvironment(t)
	t.Setenv("DATABASE_URL", " postgres://shop:secret@postgres:5432/shop ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DatabaseURL != "postgres://shop:secret@postgres:5432/shop" {
		t.Fatalf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.HTTPAddress != defaultHTTPAddress {
		t.Fatalf("HTTPAddress = %q, want %q", cfg.HTTPAddress, defaultHTTPAddress)
	}
	if cfg.DatabasePool.MaxConns != defaultDBMaxConns {
		t.Fatalf("MaxConns = %d, want %d", cfg.DatabasePool.MaxConns, defaultDBMaxConns)
	}
	if cfg.DatabasePool.MinConns != defaultDBMinConns {
		t.Fatalf("MinConns = %d, want %d", cfg.DatabasePool.MinConns, defaultDBMinConns)
	}
}

func TestLoadCustomValues(t *testing.T) {
	clearOptionalEnvironment(t)
	t.Setenv("DATABASE_URL", "postgres://shop:secret@postgres:5432/shop")
	t.Setenv("HTTP_ADDRESS", "127.0.0.1:8080")
	t.Setenv("HTTP_READ_TIMEOUT", "20s")
	t.Setenv("DB_MAX_CONNS", "40")
	t.Setenv("DB_MIN_CONNS", "8")
	t.Setenv("DB_CONNECT_TIMEOUT", "3s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPAddress != "127.0.0.1:8080" {
		t.Fatalf("HTTPAddress = %q", cfg.HTTPAddress)
	}
	if cfg.HTTPReadTimeout != 20*time.Second {
		t.Fatalf("HTTPReadTimeout = %s", cfg.HTTPReadTimeout)
	}
	if cfg.DatabasePool.MaxConns != 40 || cfg.DatabasePool.MinConns != 8 {
		t.Fatalf("pool limits = %d/%d", cfg.DatabasePool.MinConns, cfg.DatabasePool.MaxConns)
	}
	if cfg.DatabasePool.ConnectTimeout != 3*time.Second {
		t.Fatalf("ConnectTimeout = %s", cfg.DatabasePool.ConnectTimeout)
	}
}

func TestLoadValidation(t *testing.T) {
	tests := []struct {
		name      string
		variable  string
		value     string
		errorText string
	}{
		{
			name:      "database URL is required",
			variable:  "DATABASE_URL",
			value:     "",
			errorText: "DATABASE_URL",
		},
		{
			name:      "duration must be valid",
			variable:  "HTTP_READ_TIMEOUT",
			value:     "invalid",
			errorText: "HTTP_READ_TIMEOUT",
		},
		{
			name:      "connection count must be an integer",
			variable:  "DB_MAX_CONNS",
			value:     "many",
			errorText: "DB_MAX_CONNS",
		},
		{
			name:      "minimum cannot exceed maximum",
			variable:  "DB_MIN_CONNS",
			value:     "26",
			errorText: "pool limits",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clearOptionalEnvironment(t)
			t.Setenv("DATABASE_URL", "postgres://shop:secret@postgres:5432/shop")
			t.Setenv(test.variable, test.value)

			cfg, err := Load()
			if err == nil {
				t.Fatalf("Load() config = %#v, want an error", cfg)
			}
			if !strings.Contains(err.Error(), test.errorText) {
				t.Fatalf("Load() error = %q, want it to contain %q", err, test.errorText)
			}
		})
	}
}

func clearOptionalEnvironment(t *testing.T) {
	t.Helper()

	for _, name := range optionalEnvironment {
		t.Setenv(name, "")
	}
}
