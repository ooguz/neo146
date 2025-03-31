package config

import (
	"os"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	// Save original env vars
	originalEnv := os.Getenv("APP_ENV")
	originalUsername := os.Getenv("SMS_USERNAME")
	originalPassword := os.Getenv("SMS_PASSWORD")
	originalSourceAddr := os.Getenv("SMS_SOURCE_ADDR")

	// Restore env vars after test
	defer func() {
		os.Setenv("APP_ENV", originalEnv)
		os.Setenv("SMS_USERNAME", originalUsername)
		os.Setenv("SMS_PASSWORD", originalPassword)
		os.Setenv("SMS_SOURCE_ADDR", originalSourceAddr)
	}()

	testCases := []struct {
		name         string
		envVars      map[string]string
		expected     Environment
		expectSMSCfg bool
	}{
		{
			name: "Production environment",
			envVars: map[string]string{
				"APP_ENV":         "prod",
				"SMS_USERNAME":    "testuser",
				"SMS_PASSWORD":    "testpass",
				"SMS_SOURCE_ADDR": "test-source",
			},
			expected:     EnvProd,
			expectSMSCfg: true,
		},
		{
			name: "Test environment",
			envVars: map[string]string{
				"APP_ENV":         "test",
				"SMS_USERNAME":    "testuser",
				"SMS_PASSWORD":    "testpass",
				"SMS_SOURCE_ADDR": "test-source",
			},
			expected:     EnvTest,
			expectSMSCfg: true,
		},
		{
			name:         "Default environment (no env vars)",
			envVars:      map[string]string{},
			expected:     EnvProd, // Should default to prod
			expectSMSCfg: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear any existing env vars first
			os.Unsetenv("APP_ENV")
			os.Unsetenv("SMS_USERNAME")
			os.Unsetenv("SMS_PASSWORD")
			os.Unsetenv("SMS_SOURCE_ADDR")

			// Setup env vars for this test case
			for k, v := range tc.envVars {
				os.Setenv(k, v)
			}

			// Create new config
			cfg, err := NewConfig()
			if err != nil {
				t.Fatalf("Failed to create config: %v", err)
			}

			// Check environment
			if cfg.Environment != tc.expected {
				t.Errorf("Expected environment %s, got %s", tc.expected, cfg.Environment)
			}

			// Check HTTP client configuration
			if cfg.HTTPClient == nil {
				t.Errorf("Expected HTTPClient to be configured, got nil")
			}
			if cfg.HTTPClient.Timeout != 10*time.Second {
				t.Errorf("Expected 10s timeout, got %v", cfg.HTTPClient.Timeout)
			}

			// Check SMS config
			if tc.expectSMSCfg {
				if cfg.SMSUsername != tc.envVars["SMS_USERNAME"] {
					t.Errorf("Expected SMS username %s, got %s", tc.envVars["SMS_USERNAME"], cfg.SMSUsername)
				}
				if cfg.SMSPassword != tc.envVars["SMS_PASSWORD"] {
					t.Errorf("Expected SMS password %s, got %s", tc.envVars["SMS_PASSWORD"], cfg.SMSPassword)
				}
				if cfg.SMSSourceAddr != tc.envVars["SMS_SOURCE_ADDR"] {
					t.Errorf("Expected SMS source addr %s, got %s", tc.envVars["SMS_SOURCE_ADDR"], cfg.SMSSourceAddr)
				}
			}

			// Check OpenAPI spec
			if cfg.OpenAPISpec == "" {
				t.Errorf("Expected OpenAPI spec to be populated")
			}
		})
	}
}
