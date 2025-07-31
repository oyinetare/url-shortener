package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Test default values
	cfg := LoadConfig()
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "127.0.0.1", cfg.DB.Host)
	assert.Equal(t, 3306, cfg.DB.Port)
	assert.Equal(t, "urls", cfg.DB.Database)
	assert.Equal(t, "url_shorten_service", cfg.DB.User)
	assert.Equal(t, "123", cfg.DB.Password)
}

func TestNewWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("PORT", "9000")
	os.Setenv("DATABASE_HOST", "db.example.com")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_HOST")
	}()

	cfg := LoadConfig()
	assert.Equal(t, 9000, cfg.Port)
	assert.Equal(t, "db.example.com", cfg.DB.Host)
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "returns env value when set",
			envKey:       "TEST_ENV",
			envValue:     "test_value",
			defaultValue: "default",
			expected:     "test_value",
		},
		{
			name:         "returns default when env not set",
			envKey:       "UNSET_ENV",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnv(tt.envKey, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue int
		expected     int
	}{
		{
			name:         "returns int value when valid",
			envValue:     "42",
			defaultValue: 10,
			expected:     42,
		},
		{
			name:         "returns default when invalid",
			envValue:     "not_a_number",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "returns default when empty",
			envValue:     "",
			defaultValue: 10,
			expected:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("TEST_INT", tt.envValue)
				defer os.Unsetenv("TEST_INT")
			}

			result := getEnvAsInt("TEST_INT", tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}
