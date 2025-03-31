package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
)

func TestVerifyBuyMeACoffeeWebhook(t *testing.T) {
	// Set up test environment
	testSecret := "test_secret"
	os.Setenv("BUYMEACOFFEE_WEBHOOK_SECRET", testSecret)
	defer os.Unsetenv("BUYMEACOFFEE_WEBHOOK_SECRET")

	// Test payload
	testPayload := []byte(`{"type":"coffee_purchase","data":{"supporter_email":"test@example.com"}}`)

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write(testPayload)
	expectedSignature := fmt.Sprintf("%x", mac.Sum(nil))

	// Test with valid signature
	err := VerifyBuyMeACoffeeWebhook(testPayload, expectedSignature)
	if err != nil {
		t.Errorf("Expected no error for valid signature, got: %v", err)
	}

	// Test with invalid signature
	err = VerifyBuyMeACoffeeWebhook(testPayload, "invalid_signature")
	if err == nil {
		t.Errorf("Expected error for invalid signature, got nil")
	}

	// Test with missing environment variable
	os.Unsetenv("BUYMEACOFFEE_WEBHOOK_SECRET")
	err = VerifyBuyMeACoffeeWebhook(testPayload, expectedSignature)
	if err == nil {
		t.Errorf("Expected error for missing env variable, got nil")
	}
}
