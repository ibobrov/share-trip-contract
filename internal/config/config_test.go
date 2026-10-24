package config

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func validEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"HTTP_PORT",
		"REQUEST_TIMEOUT_MS",
		"LOG_LEVEL",
		"DATABASE_DSN",
		"LOG_FILE",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_DB",
		"POSTGRES_PORT",
		"DB_HOST",
		"DB_SSLMODE",
	} {
		t.Setenv(key, "")
	}
	t.Setenv("DATABASE_DSN", "postgres://test-user:test-password@localhost:5432/testdb?sslmode=disable")
}

func TestLoad_ValidConfiguration(t *testing.T) {
	validEnvironment(t)
	t.Setenv("LOG_LEVEL", "DEBUG")
	t.Setenv("REQUEST_TIMEOUT_MS", "3000")
	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, 3*time.Second, cfg.RequestTimeout)
	require.Equal(t, slog.LevelDebug, cfg.LogLevel)
	require.Equal(t, "4002", cfg.HTTPPort)
}

func TestLoad_InvalidConfiguration(t *testing.T) {
	tests := []struct{ name, key, value, want string }{
		{"missing database", "DATABASE_DSN", "", "DATABASE_DSN"},
		{"invalid database", "DATABASE_DSN", "postgres://user:must-not-appear@localhost:not-a-port/database", "DATABASE_DSN"},
		{"invalid HTTP port", "HTTP_PORT", "text", "HTTP_PORT"},
		{"HTTP port out of range", "HTTP_PORT", "65536", "HTTP_PORT"},
		{"invalid timeout", "REQUEST_TIMEOUT_MS", "abc", "REQUEST_TIMEOUT_MS"},
		{"negative timeout", "REQUEST_TIMEOUT_MS", "-1", "REQUEST_TIMEOUT_MS"},
		{"duration overflow", "REQUEST_TIMEOUT_MS", "9223372036854775807", "REQUEST_TIMEOUT_MS"},
		{"invalid log level", "LOG_LEVEL", "unknown", "LOG_LEVEL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validEnvironment(t)
			t.Setenv(tt.key, tt.value)
			_, err := Load()
			require.ErrorContains(t, err, tt.want)
			require.NotContains(t, err.Error(), "must-not-appear")
			require.NotContains(t, err.Error(), "test-password")
			require.NotContains(t, err.Error(), "test-keycloak-secret")
		})
	}
}

func TestLoad_LocalDatabaseVariables(t *testing.T) {
	validEnvironment(t)
	t.Setenv("DATABASE_DSN", "")
	t.Setenv("POSTGRES_USER", "local-user")
	t.Setenv("POSTGRES_PASSWORD", "password@with:special/characters")
	t.Setenv("POSTGRES_DB", "local-database")
	cfg, err := Load()
	require.NoError(t, err)
	require.Contains(t, cfg.DatabaseDSN, "local-database")
	require.NotContains(t, cfg.DatabaseDSN, "password@with:special/characters")
	t.Setenv("POSTGRES_PASSWORD", "")
	_, err = Load()
	require.ErrorContains(t, err, "POSTGRES_PASSWORD")
}
