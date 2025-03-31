package controllers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// TestWebhookSubscriptionInterface interface for testing
type TestWebhookSubscriptionInterface interface {
	SaveSubscription(subscriptionID string, email string, status string, expiryDate interface{}) error
	UpdateSubscriptionStatus(subscriptionID string, status string) error
	LinkPhoneToSubscription(phoneNumber string, email string) error
	UpdateRateLimitForPhone(phoneNumber string, limit int) error
}

// TestWebhookController test controller with interface
type TestWebhookController struct {
	subscriptionService TestWebhookSubscriptionInterface
}

// NewTestWebhookController creates a test webhook controller
func NewTestWebhookController(
	subscriptionService TestWebhookSubscriptionInterface,
) *TestWebhookController {
	return &TestWebhookController{
		subscriptionService: subscriptionService,
	}
}

// HandleBuyMeACoffeeWebhook handles BuyMeACoffee webhook callbacks
func (c *TestWebhookController) HandleBuyMeACoffeeWebhook(ctx *fiber.Ctx) error {
	// Parse webhook payload
	var payload map[string]interface{}
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(400).SendString(err.Error())
	}

	// Simple validation
	if payload["type"] == nil || payload["response"] == nil {
		return ctx.Status(400).SendString("Invalid webhook payload")
	}

	webhookType, ok := payload["type"].(string)
	if !ok {
		return ctx.Status(400).SendString("Invalid webhook type")
	}

	response, ok := payload["response"].(map[string]interface{})
	if !ok {
		return ctx.Status(400).SendString("Invalid webhook response")
	}

	// Handle different webhook types
	switch webhookType {
	case "subscription_created", "subscription_updated":
		subscriptionID, ok := response["subscription_id"].(string)
		if !ok {
			return ctx.Status(400).SendString("Invalid subscription_id")
		}

		email, ok := response["supporter_email"].(string)
		if !ok {
			return ctx.Status(400).SendString("Invalid supporter_email")
		}

		status, ok := response["status"].(string)
		if !ok {
			return ctx.Status(400).SendString("Invalid status")
		}

		// Mock the expiry date calculation
		expiryDate := time.Now().Add(30 * 24 * time.Hour)

		// Save subscription
		if err := c.subscriptionService.SaveSubscription(subscriptionID, email, status, expiryDate); err != nil {
			return ctx.Status(500).SendString("Error saving subscription")
		}

		// Process phone linkage if available
		if phone, ok := response["phone"].(string); ok && phone != "" {
			if err := c.subscriptionService.LinkPhoneToSubscription(phone, email); err != nil {
				return ctx.Status(500).SendString("Error linking phone to subscription")
			}

			// Set appropriate rate limit based on subscription tier
			var rateLimit int = 5 // Default rate limit
			if tier, ok := response["tier"].(string); ok {
				switch tier {
				case "basic":
					rateLimit = 10
				case "premium":
					rateLimit = 20
				case "enterprise":
					rateLimit = 50
				}
			}

			if err := c.subscriptionService.UpdateRateLimitForPhone(phone, rateLimit); err != nil {
				return ctx.Status(500).SendString("Error updating rate limit")
			}
		}

	case "subscription_cancelled", "subscription_deleted":
		subscriptionID, ok := response["subscription_id"].(string)
		if !ok {
			return ctx.Status(400).SendString("Invalid subscription_id")
		}

		if err := c.subscriptionService.UpdateSubscriptionStatus(subscriptionID, "cancelled"); err != nil {
			return ctx.Status(500).SendString("Error updating subscription status")
		}

	default:
		return ctx.Status(400).SendString("Unsupported webhook type")
	}

	return ctx.Status(200).SendString("Webhook processed successfully")
}

// HandleLinkPhoneToSubscription handles linking a phone number to a subscription
func (c *TestWebhookController) HandleLinkPhoneToSubscription(ctx *fiber.Ctx) error {
	// Parse request
	var request struct {
		Phone string `json:"phone"`
		Email string `json:"email"`
	}

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(400).SendString(err.Error())
	}

	// Validate request
	if request.Phone == "" || request.Email == "" {
		return ctx.Status(400).SendString("Phone and email are required")
	}

	// Link phone to subscription
	if err := c.subscriptionService.LinkPhoneToSubscription(request.Phone, request.Email); err != nil {
		return ctx.Status(500).SendString("Error linking phone to subscription")
	}

	// Set default rate limit
	if err := c.subscriptionService.UpdateRateLimitForPhone(request.Phone, 10); err != nil {
		return ctx.Status(500).SendString("Error updating rate limit")
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Phone linked to subscription successfully",
	})
}

// MockSubscriptionServiceForWebhook mock service for webhook tests
type MockSubscriptionServiceForWebhook struct {
	saveSubscriptionCalled        bool
	updateSubscriptionCalled      bool
	linkPhoneToSubscriptionCalled bool
	updateRateLimitCalled         bool
	shouldError                   bool
}

func NewMockSubscriptionServiceForWebhook(shouldError bool) *MockSubscriptionServiceForWebhook {
	return &MockSubscriptionServiceForWebhook{
		shouldError: shouldError,
	}
}

func (m *MockSubscriptionServiceForWebhook) SaveSubscription(subscriptionID string, email string, status string, expiryDate interface{}) error {
	m.saveSubscriptionCalled = true
	if m.shouldError {
		return fiber.ErrInternalServerError
	}
	return nil
}

func (m *MockSubscriptionServiceForWebhook) UpdateSubscriptionStatus(subscriptionID string, status string) error {
	m.updateSubscriptionCalled = true
	if m.shouldError {
		return fiber.ErrInternalServerError
	}
	return nil
}

func (m *MockSubscriptionServiceForWebhook) LinkPhoneToSubscription(phoneNumber string, email string) error {
	m.linkPhoneToSubscriptionCalled = true
	if m.shouldError {
		return fiber.ErrInternalServerError
	}
	return nil
}

func (m *MockSubscriptionServiceForWebhook) UpdateRateLimitForPhone(phoneNumber string, limit int) error {
	m.updateRateLimitCalled = true
	if m.shouldError {
		return fiber.ErrInternalServerError
	}
	return nil
}

func setupWebhookTestApp() (*fiber.App, *MockSubscriptionServiceForWebhook) {
	app := fiber.New()

	subscriptionService := NewMockSubscriptionServiceForWebhook(false)
	controller := NewTestWebhookController(subscriptionService)

	app.Post("/webhooks/buymeacoffee", controller.HandleBuyMeACoffeeWebhook)
	app.Post("/link-phone", controller.HandleLinkPhoneToSubscription)

	return app, subscriptionService
}

func TestWebhookController_HandleBuyMeACoffeeWebhook_SubscriptionCreated(t *testing.T) {
	app, subscriptionService := setupWebhookTestApp()

	// Test subscription_created webhook
	payload := map[string]interface{}{
		"type": "subscription_created",
		"response": map[string]interface{}{
			"subscription_id": "sub_123",
			"supporter_email": "test@example.com",
			"status":          "active",
			"phone":           "+1234567890",
			"tier":            "premium",
		},
	}

	jsonData, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhooks/buymeacoffee", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error testing request: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify that methods were called
	if !subscriptionService.saveSubscriptionCalled {
		t.Error("Expected SaveSubscription to be called")
	}

	if !subscriptionService.linkPhoneToSubscriptionCalled {
		t.Error("Expected LinkPhoneToSubscription to be called")
	}

	if !subscriptionService.updateRateLimitCalled {
		t.Error("Expected UpdateRateLimitForPhone to be called")
	}
}

func TestWebhookController_HandleBuyMeACoffeeWebhook_SubscriptionCancelled(t *testing.T) {
	app, subscriptionService := setupWebhookTestApp()

	// Test subscription_cancelled webhook
	payload := map[string]interface{}{
		"type": "subscription_cancelled",
		"response": map[string]interface{}{
			"subscription_id": "sub_123",
		},
	}

	jsonData, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhooks/buymeacoffee", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error testing request: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify that methods were called
	if !subscriptionService.updateSubscriptionCalled {
		t.Error("Expected UpdateSubscriptionStatus to be called")
	}
}

func TestWebhookController_HandleBuyMeACoffeeWebhook_InvalidPayload(t *testing.T) {
	app, _ := setupWebhookTestApp()

	// Test invalid payload
	payload := map[string]interface{}{
		"invalid": "payload",
	}

	jsonData, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhooks/buymeacoffee", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error testing request: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestWebhookController_HandleBuyMeACoffeeWebhook_ServiceError(t *testing.T) {
	app := fiber.New()

	// Create service that will return errors
	subscriptionService := NewMockSubscriptionServiceForWebhook(true)
	controller := NewTestWebhookController(subscriptionService)

	app.Post("/webhooks/buymeacoffee", controller.HandleBuyMeACoffeeWebhook)

	// Test subscription_created webhook with service error
	payload := map[string]interface{}{
		"type": "subscription_created",
		"response": map[string]interface{}{
			"subscription_id": "sub_123",
			"supporter_email": "test@example.com",
			"status":          "active",
		},
	}

	jsonData, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhooks/buymeacoffee", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error testing request: %v", err)
	}

	if resp.StatusCode != 500 {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
}

func TestWebhookController_HandleLinkPhoneToSubscription(t *testing.T) {
	app, subscriptionService := setupWebhookTestApp()

	// Test link phone request
	payload := map[string]interface{}{
		"phone": "+1234567890",
		"email": "test@example.com",
	}

	jsonData, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/link-phone", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error testing request: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify that methods were called
	if !subscriptionService.linkPhoneToSubscriptionCalled {
		t.Error("Expected LinkPhoneToSubscription to be called")
	}

	if !subscriptionService.updateRateLimitCalled {
		t.Error("Expected UpdateRateLimitForPhone to be called")
	}

	// Test invalid request
	payload = map[string]interface{}{
		"phone": "",
		"email": "test@example.com",
	}

	jsonData, _ = json.Marshal(payload)
	req = httptest.NewRequest("POST", "/link-phone", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("Error testing request: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}
