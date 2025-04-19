package controllers

import (
	"fmt"
	"neo146/models"
	"neo146/services"
	"neo146/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

// WebhookController handles webhook-related endpoints
type WebhookController struct {
	subscriptionService *services.SubscriptionService
}

// NewWebhookController creates a new WebhookController
func NewWebhookController(subscriptionService *services.SubscriptionService) *WebhookController {
	return &WebhookController{
		subscriptionService: subscriptionService,
	}
}

// HandleBuyMeACoffee handles the Buy Me a Coffee webhook
func (c *WebhookController) HandleBuyMeACoffee(ctx *fiber.Ctx) error {
	// Get the raw body for verification
	body := ctx.Body()

	// Get the signature from headers
	fmt.Println("--------------------------------")
	fmt.Println("Headers:")
	ctx.Request().Header.VisitAll(func(key, value []byte) {
		fmt.Printf("%s: %s\n", string(key), string(value))
	})
	fmt.Println("X-Bmc-Signature: ", ctx.Get("X-Bmc-Signature"))
	fmt.Println("X-BMC-Signature: ", ctx.Get("X-BMC-Signature"))
	fmt.Println("Body: ", string(body))
	fmt.Println("--------------------------------")
	signature := ctx.Get("X-Signature-Sha256")

	// Verify the webhook signature
	if err := utils.VerifyBuyMeACoffeeWebhook(body, signature); err != nil {
		return ctx.Status(400).SendString(err.Error())
	}

	var webhook models.BuyMeACoffeeWebhook
	if err := ctx.BodyParser(&webhook); err != nil {
		return ctx.Status(400).SendString(err.Error())
	}

	// Handle different webhook types
	switch webhook.Type {
	case "recurring_donation.started":
		// Save subscription with active status
		if err := c.subscriptionService.SaveSubscription(
			webhook.Data.PspID,
			webhook.Data.SupporterEmail,
			"active",
			time.Unix(webhook.Data.CurrentPeriodEnd, 0),
		); err != nil {
			return ctx.Status(500).SendString("Error saving subscription: " + err.Error())
		}

	case "recurring_donation.updated":
		// Check if subscription is paused or canceled
		if webhook.Data.Paused == "true" || webhook.Data.Canceled == "true" {
			if err := c.subscriptionService.UpdateSubscriptionStatus(webhook.Data.PspID, "inactive"); err != nil {
				return ctx.Status(500).SendString("Error updating subscription status: " + err.Error())
			}
		}

	case "recurring_donation.cancelled":
		// Update subscription status to cancelled
		if err := c.subscriptionService.UpdateSubscriptionStatus(webhook.Data.PspID, "cancelled"); err != nil {
			return ctx.Status(500).SendString("Error updating subscription status: " + err.Error())
		}
	}

	return ctx.SendStatus(200)
}

// HandlePayPal handles the PayPal IPN webhook
func (c *WebhookController) HandlePayPal(ctx *fiber.Ctx) error {
	// Get the raw body for verification
	body := ctx.Body()

	// Parse the form data
	form, err := ctx.MultipartForm()
	if err != nil {
		return ctx.Status(400).SendString("Invalid form data")
	}

	// Create IPN struct
	ipn := &models.PayPalIPN{
		PaymentStatus:    form.Value["payment_status"][0],
		PaymentType:      form.Value["payment_type"][0],
		PaymentDate:      form.Value["payment_date"][0],
		PaymentGross:     utils.ParseFloat(form.Value["mc_gross"][0]),
		PaymentFee:       utils.ParseFloat(form.Value["mc_fee"][0]),
		Currency:         form.Value["mc_currency"][0],
		PayerEmail:       form.Value["payer_email"][0],
		PayerID:          form.Value["payer_id"][0],
		SubscriptionID:   form.Value["subscr_id"][0],
		SubscriptionDate: form.Value["subscr_date"][0],
		SubscriptionEnd:  form.Value["subscr_end"][0],
		Custom:           form.Value["custom"][0],
		IPNType:          form.Value["txn_type"][0],
	}

	// Verify the IPN with PayPal
	if err := utils.VerifyPayPalIPN(body); err != nil {
		return ctx.Status(400).SendString("Invalid IPN")
	}

	// Handle different IPN types
	switch ipn.IPNType {
	case "subscr_signup":
		// New subscription
		if err := c.subscriptionService.SaveSubscription(
			ipn.SubscriptionID,
			ipn.PayerEmail,
			"active",
			utils.ParsePayPalDate(ipn.SubscriptionEnd),
		); err != nil {
			return ctx.Status(500).SendString("Error saving subscription")
		}

	case "subscr_payment":
		// Subscription payment received
		if err := c.subscriptionService.UpdateSubscriptionStatus(ipn.SubscriptionID, "active"); err != nil {
			return ctx.Status(500).SendString("Error updating subscription")
		}

	case "subscr_cancel":
		// Subscription cancelled
		if err := c.subscriptionService.UpdateSubscriptionStatus(ipn.SubscriptionID, "cancelled"); err != nil {
			return ctx.Status(500).SendString("Error updating subscription")
		}

	case "subscr_eot":
		// Subscription ended
		if err := c.subscriptionService.UpdateSubscriptionStatus(ipn.SubscriptionID, "expired"); err != nil {
			return ctx.Status(500).SendString("Error updating subscription")
		}
	}

	// Send success response back to PayPal
	return ctx.SendString("OK")
}
