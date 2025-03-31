package services

import (
	"testing"
	"time"
)

func TestSubscriptionService_SaveSubscription(t *testing.T) {
	service := NewSubscriptionService()

	// Test a basic subscription save
	err := service.SaveSubscription(
		"sub123",
		"test@example.com",
		"active",
		time.Now().Add(30*24*time.Hour),
	)

	if err != nil {
		t.Errorf("Expected successful subscription save, got error: %v", err)
	}

	// Since this is a mock implementation that just logs, we can't check the actual
	// data storage, but we can verify it doesn't return an error
}

func TestSubscriptionService_UpdateSubscriptionStatus(t *testing.T) {
	service := NewSubscriptionService()

	// Test status update
	err := service.UpdateSubscriptionStatus("sub123", "canceled")
	if err != nil {
		t.Errorf("Expected successful status update, got error: %v", err)
	}
}

func TestSubscriptionService_LinkPhoneToSubscription(t *testing.T) {
	service := NewSubscriptionService()

	// Test phone linking
	err := service.LinkPhoneToSubscription("+1234567890", "test@example.com")
	if err != nil {
		t.Errorf("Expected successful phone linking, got error: %v", err)
	}
}

func TestSubscriptionService_UpdateRateLimitForPhone(t *testing.T) {
	service := NewSubscriptionService()

	// Test rate limit update
	err := service.UpdateRateLimitForPhone("+1234567890", 100)
	if err != nil {
		t.Errorf("Expected successful rate limit update, got error: %v", err)
	}
}

func TestSubscriptionService_CheckRateLimit(t *testing.T) {
	service := NewSubscriptionService()

	// Test rate limit check
	allowed, err := service.CheckRateLimit("+1234567890")
	if err != nil {
		t.Errorf("Expected successful rate limit check, got error: %v", err)
	}

	// The mock implementation always returns true
	if !allowed {
		t.Error("Expected rate limit check to allow, but it did not")
	}
}
