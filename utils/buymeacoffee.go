package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"os"
)

// VerifyBuyMeACoffeeWebhook verifies a BuyMeACoffee webhook signature
func VerifyBuyMeACoffeeWebhook(body []byte, signature string) error {
	// Get the secret from environment
	secret := os.Getenv("BUYMEACOFFEE_WEBHOOK_SECRET")
	if secret == "" {
		return fmt.Errorf("BUYMEACOFFEE_WEBHOOK_SECRET environment variable not set")
	}

	// Calculate HMAC SHA256 signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := fmt.Sprintf("%x", mac.Sum(nil))

	if signature != expectedSignature {
		return fmt.Errorf("invalid BuyMeACoffee webhook signature")
	}

	return nil
}
