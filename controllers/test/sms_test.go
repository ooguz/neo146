package test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"smsgw/controllers"
	"smsgw/models"
	"smsgw/providers"
	"smsgw/services"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestSMSController_HandleTest_ProductionEnv(t *testing.T) {
	// Create a new fiber app
	app := fiber.New()

	// Set up SMS manager
	smsManager := providers.NewManager()

	// Create services
	smsService := services.NewSMSService(smsManager)
	markdownService := services.NewMarkdownService(nil)
	twitterService := services.NewTwitterService(nil)
	searchService := services.NewSearchService(nil)
	weatherService := services.NewWeatherService(nil)
	subscriptionService := services.NewSubscriptionService()

	// Create SMS controller with production environment
	controller := controllers.NewSMSController(
		&controllers.Config{Environment: "prod"},
		smsService,
		markdownService,
		twitterService,
		searchService,
		weatherService,
		subscriptionService,
	)

	// Setup the route
	app.Post("/test", controller.HandleTest)

	// Create a test request
	payload := []models.SMSPayload{
		{SourceAddr: "+1234567890", Content: "Test message"},
	}
	jsonData, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, _ := app.Test(req)

	// Assert the response
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSMSController_HandleTest_TestEnv(t *testing.T) {
	// Save and restore env vars
	origUsername := os.Getenv("SMS_USERNAME")
	origPassword := os.Getenv("SMS_PASSWORD")
	origSourceAddr := os.Getenv("SMS_SOURCE_ADDR")
	defer func() {
		os.Setenv("SMS_USERNAME", origUsername)
		os.Setenv("SMS_PASSWORD", origPassword)
		os.Setenv("SMS_SOURCE_ADDR", origSourceAddr)
	}()

	// Set env vars for test
	os.Setenv("SMS_USERNAME", "testuser")
	os.Setenv("SMS_PASSWORD", "testpass")
	os.Setenv("SMS_SOURCE_ADDR", "testsource")

	// Create a new fiber app
	app := fiber.New()

	// Set up SMS manager
	smsManager := providers.NewManager()

	// Create services
	smsService := services.NewSMSService(smsManager)
	markdownService := services.NewMarkdownService(nil)
	twitterService := services.NewTwitterService(nil)
	searchService := services.NewSearchService(nil)
	weatherService := services.NewWeatherService(nil)
	subscriptionService := services.NewSubscriptionService()

	// Create SMS controller with test environment
	controller := controllers.NewSMSController(
		&controllers.Config{Environment: "test"},
		smsService,
		markdownService,
		twitterService,
		searchService,
		weatherService,
		subscriptionService,
	)

	// Setup the route
	app.Post("/test", controller.HandleTest)

	// Create a test request
	payload := []models.SMSPayload{
		{SourceAddr: "+1234567890", Content: "Test message"},
	}
	jsonData, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, _ := app.Test(req)

	// Assert the response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
