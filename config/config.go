package config

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Environment type
type Environment string

const (
	EnvTest Environment = "test"
	EnvProd Environment = "prod"
)

//go:embed openapi.json
var openAPISpec string

// Config holds all configuration for the application
type Config struct {
	Environment   Environment
	HTTPClient    *http.Client
	SMSUsername   string
	SMSPassword   string
	SMSSourceAddr string
	OpenAPISpec   string
}

// NewConfig creates a new Config instance
func NewConfig() (*Config, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: Error loading .env file")
	}

	// Set environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "prod" // Default to production if not set
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &Config{
		Environment:   Environment(env),
		HTTPClient:    httpClient,
		SMSUsername:   os.Getenv("SMS_USERNAME"),
		SMSPassword:   os.Getenv("SMS_PASSWORD"),
		SMSSourceAddr: os.Getenv("SMS_SOURCE_ADDR"),
		OpenAPISpec:   openAPISpec,
	}, nil
}
