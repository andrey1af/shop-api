package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/andrey1af/shop-api/api-service/internal/database"
)

const (
	defaultHTTPAddress           = ":8090"
	defaultHTTPReadHeaderTimeout = 5 * time.Second
	defaultHTTPReadTimeout       = 15 * time.Second
	defaultHTTPWriteTimeout      = 30 * time.Second
	defaultHTTPIdleTimeout       = 60 * time.Second
	defaultHTTPShutdownTimeout   = 5 * time.Second
	defaultDBMaxConns            = int32(25)
	defaultDBMinConns            = int32(5)
	defaultDBMaxConnLifetime     = 30 * time.Minute
	defaultDBMaxConnIdleTime     = 5 * time.Minute
	defaultDBHealthCheckPeriod   = time.Minute
	defaultDBConnectTimeout      = 5 * time.Second
)

type Config struct {
	HTTPAddress           string
	HTTPReadHeaderTimeout time.Duration
	HTTPReadTimeout       time.Duration
	HTTPWriteTimeout      time.Duration
	HTTPIdleTimeout       time.Duration
	HTTPShutdownTimeout   time.Duration
	DatabaseURL           string
	DatabasePool          database.PoolConfig
}

func Load() (*Config, error) {
	httpReadHeaderTimeout, err := envDuration("HTTP_READ_HEADER_TIMEOUT", defaultHTTPReadHeaderTimeout)
	if err != nil {
		return nil, err
	}

	httpReadTimeout, err := envDuration("HTTP_READ_TIMEOUT", defaultHTTPReadTimeout)
	if err != nil {
		return nil, err
	}

	httpWriteTimeout, err := envDuration("HTTP_WRITE_TIMEOUT", defaultHTTPWriteTimeout)
	if err != nil {
		return nil, err
	}

	httpIdleTimeout, err := envDuration("HTTP_IDLE_TIMEOUT", defaultHTTPIdleTimeout)
	if err != nil {
		return nil, err
	}

	httpShutdownTimeout, err := envDuration("HTTP_SHUTDOWN_TIMEOUT", defaultHTTPShutdownTimeout)
	if err != nil {
		return nil, err
	}

	databasePool, err := loadDatabasePool()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		HTTPAddress:           env("HTTP_ADDRESS", defaultHTTPAddress),
		HTTPReadHeaderTimeout: httpReadHeaderTimeout,
		HTTPReadTimeout:       httpReadTimeout,
		HTTPWriteTimeout:      httpWriteTimeout,
		HTTPIdleTimeout:       httpIdleTimeout,
		HTTPShutdownTimeout:   httpShutdownTimeout,
		DatabaseURL:           strings.TrimSpace(os.Getenv("DATABASE_URL")),
		DatabasePool:          databasePool,
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	if cfg.HTTPShutdownTimeout <= 0 {
		return nil, errors.New("HTTP_SHUTDOWN_TIMEOUT must be greater than zero")
	}

	return cfg, nil
}

func loadDatabasePool() (database.PoolConfig, error) {
	maxConns, err := envInt32("DB_MAX_CONNS", defaultDBMaxConns)
	if err != nil {
		return database.PoolConfig{}, err
	}

	minConns, err := envInt32("DB_MIN_CONNS", defaultDBMinConns)
	if err != nil {
		return database.PoolConfig{}, err
	}

	maxConnLifetime, err := envDuration("DB_MAX_CONN_LIFETIME", defaultDBMaxConnLifetime)
	if err != nil {
		return database.PoolConfig{}, err
	}

	maxConnIdleTime, err := envDuration("DB_MAX_CONN_IDLE_TIME", defaultDBMaxConnIdleTime)
	if err != nil {
		return database.PoolConfig{}, err
	}

	healthCheckPeriod, err := envDuration("DB_HEALTH_CHECK_PERIOD", defaultDBHealthCheckPeriod)
	if err != nil {
		return database.PoolConfig{}, err
	}

	connectTimeout, err := envDuration("DB_CONNECT_TIMEOUT", defaultDBConnectTimeout)
	if err != nil {
		return database.PoolConfig{}, err
	}

	if maxConns < 1 || minConns < 0 || minConns > maxConns {
		return database.PoolConfig{}, errors.New("database pool limits are invalid")
	}

	if healthCheckPeriod <= 0 || connectTimeout <= 0 {
		return database.PoolConfig{}, errors.New("database pool timeouts must be greater than zero")
	}

	return database.PoolConfig{
		MaxConns:          maxConns,
		MinConns:          minConns,
		MaxConnLifetime:   maxConnLifetime,
		MaxConnIdleTime:   maxConnIdleTime,
		HealthCheckPeriod: healthCheckPeriod,
		ConnectTimeout:    connectTimeout,
	}, nil
}

func env(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}

	return value
}

func envInt32(name string, fallback int32) (int32, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}

	return int32(parsed), nil
}

func envDuration(name string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", name, err)
	}

	return parsed, nil
}
