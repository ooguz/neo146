package controllers

import (
	"fmt"
	"log"
	"neo146/services"
	"os"
)

// TelegramController handles Telegram bot integration
type TelegramController struct {
	TelegramService *services.TelegramService
}

// NewTelegramController creates a new Telegram controller
func NewTelegramController(
	smsService *services.SMSService,
	subService *services.SubscriptionService,
) (*TelegramController, error) {
	// Get Telegram bot token from environment
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	// Create Telegram service
	telegramService, err := services.NewTelegramService(token, smsService, subService)
	if err != nil {
		return nil, fmt.Errorf("error creating Telegram service: %v", err)
	}

	return &TelegramController{
		TelegramService: telegramService,
	}, nil
}

// StartBot starts the Telegram bot
func (c *TelegramController) StartBot() error {
	log.Println("Starting Telegram bot...")
	return c.TelegramService.Start()
}

// Cleanup cleans up the bot instance
func (c *TelegramController) Cleanup() {
	if c.TelegramService != nil {
		c.TelegramService.Cleanup()
	}
}
