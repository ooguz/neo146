package controllers

import (
	"encoding/json"
	"fmt"
	"smsgw/models"
	"smsgw/providers"
	"smsgw/services"
	"smsgw/utils"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SMSController handles SMS-related endpoints
type SMSController struct {
	config              *Config
	smsService          *services.SMSService
	markdownService     *services.MarkdownService
	twitterService      *services.TwitterService
	searchService       *services.SearchService
	weatherService      *services.WeatherService
	subscriptionService *services.SubscriptionService
}

// Config holds configuration for the SMSController
type Config struct {
	Environment string
}

// NewSMSController creates a new SMSController
func NewSMSController(
	config *Config,
	smsService *services.SMSService,
	markdownService *services.MarkdownService,
	twitterService *services.TwitterService,
	searchService *services.SearchService,
	weatherService *services.WeatherService,
	subscriptionService *services.SubscriptionService,
) *SMSController {
	return &SMSController{
		config:              config,
		smsService:          smsService,
		markdownService:     markdownService,
		twitterService:      twitterService,
		searchService:       searchService,
		weatherService:      weatherService,
		subscriptionService: subscriptionService,
	}
}

// HandleTest handles the test endpoint
func (c *SMSController) HandleTest(ctx *fiber.Ctx) error {
	// Deny access in production
	if c.config.Environment == "prod" {
		return ctx.Status(403).SendString("Test endpoint is not available in production environment")
	}

	var payload []models.SMSPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(400).SendString(err.Error())
	}

	// Log received payload
	receivedJSON, _ := json.MarshalIndent(payload, "", "  ")
	fmt.Printf("Test endpoint received payload:\n%s\n", string(receivedJSON))

	// Process the payload and create response
	var response []providers.Message
	for _, sms := range payload {
		content := strings.TrimSpace(sms.Content)

		// Check if content is a URL
		if utils.IsURL(content) {
			markdown, err := c.markdownService.FetchMarkdown(content)
			if err != nil {
				fmt.Println("Error fetching markdown:", err)
				continue
			}

			// Split and encode the message
			encodedParts := utils.SplitAndEncodeMessage(markdown, 500)

			// Create response messages
			for i, encoded := range encodedParts {
				response = append(response, providers.Message{
					Msg:  encoded,
					Dest: sms.SourceAddr,
					ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), i),
				})
			}
			continue
		}

		// Check if content matches "twitter user <username>"
		if strings.HasPrefix(strings.ToLower(content), "twitter user ") {
			username := strings.TrimSpace(strings.TrimPrefix(content, "twitter user"))
			tweets, err := c.twitterService.FetchTweets(username, 5)
			if err != nil {
				fmt.Println("Error fetching tweets:", err)
				continue
			}

			// Split and encode the message
			encodedParts := utils.SplitAndEncodeMessage(tweets, 500)

			// Create response messages
			for i, encoded := range encodedParts {
				response = append(response, providers.Message{
					Msg:  encoded,
					Dest: sms.SourceAddr,
					ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), i),
				})
			}
			continue
		}

		// Check if content matches "websearch <query>"
		if strings.HasPrefix(strings.ToLower(content), "websearch ") {
			query := strings.TrimSpace(strings.TrimPrefix(content, "websearch"))
			results, err := c.searchService.FetchDuckDuckGoResults(query)
			if err != nil {
				fmt.Println("Error fetching search results:", err)
				continue
			}

			// Split and encode the message
			encodedParts := utils.SplitAndEncodeMessage(results, 500)

			// Create response messages
			for i, encoded := range encodedParts {
				response = append(response, providers.Message{
					Msg:  encoded,
					Dest: sms.SourceAddr,
					ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), i),
				})
			}
			continue
		}

		// Check if content matches "wiki <lang> <query>"
		if strings.HasPrefix(strings.ToLower(content), "wiki ") {
			parts := strings.SplitN(strings.TrimPrefix(content, "wiki "), " ", 2)
			if len(parts) != 2 {
				continue
			}

			langCode := strings.TrimSpace(parts[0])
			query := strings.TrimSpace(parts[1])

			summary, err := c.searchService.FetchWikipediaSummary(query, langCode)
			if err != nil {
				fmt.Println("Error fetching Wikipedia summary:", err)
				continue
			}

			// Split and encode the message
			encodedParts := utils.SplitAndEncodeMessage(summary, 500)

			// Create response messages
			for i, encoded := range encodedParts {
				response = append(response, providers.Message{
					Msg:  encoded,
					Dest: sms.SourceAddr,
					ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), i),
				})
			}
			continue
		}

		// Check if content matches "weather <location>"
		if strings.HasPrefix(strings.ToLower(content), "weather ") {
			location := strings.TrimSpace(strings.TrimPrefix(content, "weather"))
			forecast, err := c.weatherService.FetchWeatherForecast(location)
			if err != nil {
				fmt.Println("Error fetching weather forecast:", err)
				continue
			}

			// Weather is sent without encoding
			response = append(response, providers.Message{
				Msg:  forecast,
				Dest: sms.SourceAddr,
				ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), 0),
			})
			continue
		}
	}

	// Create the response payload
	responsePayload := c.smsService.CreateSMSRequest(response)
	return ctx.JSON(responsePayload)
}

// HandleInbound handles the inbound SMS endpoint
func (c *SMSController) HandleInbound(ctx *fiber.Ctx) error {
	var payload []models.SMSPayload
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(400).SendString(err.Error())
	}

	// Log received payload
	receivedJSON, _ := json.MarshalIndent(payload, "", "  ")
	fmt.Printf("Received payload:\n%s\n", string(receivedJSON))

	for _, sms := range payload {
		content := strings.TrimSpace(sms.Content)

		// Handle subscription command
		if strings.HasPrefix(strings.ToLower(content), "subscribe ") {
			email := strings.TrimSpace(strings.TrimPrefix(content, "subscribe"))
			if err := c.subscriptionService.LinkPhoneToSubscription(sms.SourceAddr, email); err != nil {
				fmt.Printf("Error linking phone to subscription: %v\n", err)
				continue
			}
			// Update rate limit for the phone number
			if err := c.subscriptionService.UpdateRateLimitForPhone(sms.SourceAddr, 20); err != nil {
				fmt.Printf("Error updating rate limit: %v\n", err)
				continue
			}
			// Send confirmation message
			confirmationMsg := "Your subscription has been activated. You now have a limit of 20 messages per hour."
			encodedParts := utils.SplitAndEncodeMessage(confirmationMsg, 500)
			for i, encoded := range encodedParts {
				smsMessage := []providers.Message{
					{
						Msg:  encoded,
						Dest: sms.SourceAddr,
						ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), i),
					},
				}
				if err := c.smsService.SendSMS(smsMessage); err != nil {
					fmt.Printf("Error sending confirmation message: %v\n", err)
				}
			}
			continue
		}

		// Check rate limit
		allowed, err := c.subscriptionService.CheckRateLimit(sms.SourceAddr)
		if err != nil {
			fmt.Printf("Error checking rate limit: %v\n", err)
			continue
		}

		if !allowed {
			// Send rate limit notification with source address
			if sms.SourceAddr == "" {
				fmt.Println("Error: source address is empty for rate limit notification")
				continue
			}

			rateLimitMsg := "!: You have reached the rate limit of 5 messages per hour. Please try again later or subscribe to the service. https://buymeacoffee.com/ooguz"
			smsMessage := []providers.Message{
				{
					Msg:  rateLimitMsg,
					Dest: sms.SourceAddr,
					ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), 0),
				},
			}
			if err := c.smsService.SendSMS(smsMessage); err != nil {
				fmt.Printf("Error sending rate limit notification: %v\n", err)
			}
			continue
		}

		// Check if content is a URL
		if utils.IsURL(content) {
			markdown, err := c.markdownService.FetchMarkdown(content)
			if err != nil {
				fmt.Println("Error fetching markdown:", err)
				continue
			}

			// Send SMS with markdown content
			if err := c.smsService.PrepareAndSendSMS(markdown, sms.SourceAddr, true); err != nil {
				fmt.Println("Error sending SMS:", err)
			}
			continue
		}

		// Check if content matches "twitter user <username>"
		if strings.HasPrefix(strings.ToLower(content), "twitter user ") {
			username := strings.TrimSpace(strings.TrimPrefix(content, "twitter user"))
			tweets, err := c.twitterService.FetchTweets(username, 5)
			if err != nil {
				fmt.Println("Error fetching tweets:", err)
				continue
			}

			// Send SMS with tweets content
			if err := c.smsService.PrepareAndSendSMS(tweets, sms.SourceAddr, true); err != nil {
				fmt.Println("Error sending SMS:", err)
			}
			continue
		}

		// Check if content matches "websearch <query>"
		if strings.HasPrefix(strings.ToLower(content), "websearch ") {
			query := strings.TrimSpace(strings.TrimPrefix(content, "websearch"))
			results, err := c.searchService.FetchDuckDuckGoResults(query)
			if err != nil {
				fmt.Println("Error fetching search results:", err)
				continue
			}

			// Send SMS with search results
			if err := c.smsService.PrepareAndSendSMS(results, sms.SourceAddr, true); err != nil {
				fmt.Println("Error sending SMS:", err)
			}
			continue
		}

		// Check if content matches "wiki <lang> <query>"
		if strings.HasPrefix(strings.ToLower(content), "wiki ") {
			parts := strings.SplitN(strings.TrimPrefix(content, "wiki "), " ", 2)
			if len(parts) != 2 {
				continue
			}

			langCode := strings.TrimSpace(parts[0])
			query := strings.TrimSpace(parts[1])

			summary, err := c.searchService.FetchWikipediaSummary(query, langCode)
			if err != nil {
				fmt.Println("Error fetching Wikipedia summary:", err)
				continue
			}

			// Send SMS with Wikipedia summary
			if err := c.smsService.PrepareAndSendSMS(summary, sms.SourceAddr, true); err != nil {
				fmt.Println("Error sending SMS:", err)
			}
			continue
		}

		// Check if content matches "weather <location>"
		if strings.HasPrefix(strings.ToLower(content), "weather ") {
			location := strings.TrimSpace(strings.TrimPrefix(content, "weather"))
			forecast, err := c.weatherService.FetchWeatherForecast(location)
			if err != nil {
				fmt.Println("Error fetching weather forecast:", err)
				continue
			}

			// Weather is sent without encoding
			if err := c.smsService.PrepareAndSendSMS(forecast, sms.SourceAddr, false); err != nil {
				fmt.Println("Error sending SMS:", err)
			}
			continue
		}
	}
	return ctx.SendStatus(204)
}

// HandleTestSubscribe handles the test subscription endpoint
func (c *SMSController) HandleTestSubscribe(ctx *fiber.Ctx) error {
	// Deny access in production
	if c.config.Environment == "prod" {
		return ctx.Status(403).SendString("Test endpoint is not available in production environment")
	}

	type SubscriptionRequest struct {
		Email string `json:"email"`
	}

	var req SubscriptionRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).SendString(err.Error())
	}

	if req.Email == "" {
		return ctx.Status(400).SendString("Email is required")
	}

	// Save subscription with active status
	if err := c.subscriptionService.SaveSubscription(
		fmt.Sprintf("test_%d", time.Now().Unix()),
		req.Email,
		"active",
		time.Now().Add(30*24*time.Hour), // 30 days subscription
	); err != nil {
		return ctx.Status(500).SendString(fmt.Sprintf("Error saving subscription: %v", err))
	}

	return ctx.SendString("Subscription added successfully")
}
