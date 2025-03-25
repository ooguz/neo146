package main

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"smsgw/providers"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

// Create HTTP client with timeout
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// Environment type
type Environment string

const (
	EnvTest Environment = "test"
	EnvProd Environment = "prod"
)

var currentEnv Environment
var smsManager *providers.Manager

type SMSPayload struct {
	MessageID       int    `json:"message_id"`
	Type            string `json:"type"`
	CreatedAt       string `json:"created_at"`
	Network         string `json:"network"`
	SourceAddr      string `json:"source_addr"`
	DestinationAddr string `json:"destination_addr"`
	Keyword         string `json:"keyword"`
	Content         string `json:"content"`
	ReceivedAt      string `json:"received_at"`
}

type SMSRequest struct {
	Username   string              `json:"username"`
	Password   string              `json:"password"`
	SourceAddr string              `json:"source_addr"`
	ValidFor   string              `json:"valid_for"`
	SendAt     string              `json:"send_at"`
	CustomID   string              `json:"custom_id"`
	Datacoding string              `json:"datacoding"`
	Messages   []providers.Message `json:"messages"`
}

type BuyMeACoffeeWebhook struct {
	Type     string `json:"type"`
	LiveMode bool   `json:"live_mode"`
	Attempt  int    `json:"attempt"`
	Created  int64  `json:"created"`
	EventID  int    `json:"event_id"`
	Data     struct {
		ID                 int     `json:"id"`
		Amount             float64 `json:"amount"`
		Object             string  `json:"object"`
		Paused             string  `json:"paused"`
		Status             string  `json:"status"`
		Canceled           string  `json:"canceled"`
		Currency           string  `json:"currency"`
		PspID              string  `json:"psp_id"`
		DurationType       string  `json:"duration_type"`
		StartedAt          int64   `json:"started_at"`
		CanceledAt         *int64  `json:"canceled_at"`
		NoteHidden         bool    `json:"note_hidden"`
		SupportNote        *string `json:"support_note"`
		SupporterName      string  `json:"supporter_name"`
		SupporterID        int     `json:"supporter_id"`
		SupporterEmail     string  `json:"supporter_email"`
		CurrentPeriodEnd   int64   `json:"current_period_end"`
		CurrentPeriodStart int64   `json:"current_period_start"`
		SupporterFeedback  *string `json:"supporter_feedback"`
		CancelAtPeriodEnd  *string `json:"cancel_at_period_end"`
	} `json:"data"`
}

// RSS feed structures
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
	}

	// Set environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "prod" // Default to production if not set
	}
	currentEnv = Environment(env)

	// Initialize database
	if err := initDB(); err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize SMS provider manager
	smsManager = providers.NewManager()
	smsManager.RegisterProvider(providers.NewVerimorProvider())

	app := fiber.New(
		fiber.Config{
			AppName:      "neo146 - smsgw",
			Prefork:      true,
			ServerHeader: "neo146-smsgw",
		},
	)

	// Add root endpoint with documentation
	app.Get("/", func(c *fiber.Ctx) error {
		doc := `neo146 - smsgw
==============

neo146 provides a minimal HTTP-SMS gateway that serves as an emergency network
connection method inspired by dial-up, allowing you to access content via SMS.

The name "neo146" comes from the Turkish historic public dial-up service, which
operated on dial number 146. 

The service is free, but running it costs about 20 EUR per month, and also ~3
cents per message for the SMS gateway. For a better experience and support the
service, please consider subscribing.

SMS responses are base64 encoded for using less SMS credits. Multiple messages
are used to send longer responses, the sequence of messages is indicated in the
response as "GW<number>|" prefix.

HTTP responses are not encoded by default, but can be requested with b64=true
parameter.

Available SMS Commands:
- URL (https://...) - Fetch and convert any webpage to Markdown format
- "twitter user <username>" - Get the last 5 tweets from a Twitter user
- "websearch <query>" - Search the web using DuckDuckGo
- "wiki <2charlangcode> <query>" - Get Wikipedia article summary
- "weather <location>" - Get weather forecast for a location

HTTP Endpoints:
- /uri2md?uri=<uri>[&b64=true] - Convert URI to Markdown
- /twitter?user=<user>[&b64=true] - Get last 5 tweets of a user
- /ddg?q=<query>[&b64=true] - Search the web via DuckDuckGo
- /wiki?lang=<2charlangcode>&q=<query>[&b64=true] - Get Wikipedia article summ.
- /weather?loc=<location> - Get weather forecast

Rate Limits:
- 5 messages per hour per phone number
- Subscribe to support the service and get 20 messages/hour per phone number

Subscription:
- Get higher rate limits by subscribing via: <https://buymeacoffee.com/ooguz>
- After subscribing, text "subscribe <your-email>" to link your phone number
- All your contribution will be used to maintain the service, rest will be 
donated to Free Software Association in Turkey (Özgür Yazılım Derneği) 
<https://oyd.org.tr>
- Running this service costs about 20 EUR per month, and also ~3 cents/message
for the SMS gateway. For a better experience and support the service, please 
consider subscribing.

Support:
- For more information or support, neo146 {at} riseup {dot} net
- E-mail is preferred for requests, but you can also contact via Twitter @ooguz
- Subscribe to our neo146-users mailing list to get updates and support:
<https://lists.riseup.net/www/subscribe/neo146-users>

Thanks:
- wttr.in for the weather data - <https://wttr.in>
- duckduckgo lite for the search engine - <https://lite.duckduckgo.com/lite>
- urltomarkdown for the md conversion - <https://github.com/macsplit/urltomarkdown>
- nitter project for the Twitter API - <https://github.com/zedeus/nitter>
- Özgür Yazılım Derneği for the support - <https://oyd.org.tr>


This gateway is free software, licensed under GNU AGPL v3 or later. 
Source code is available at:
https://github.com/ooguz/neo146

--------------------------------

Warning:
- This service is provided as-is, without any warranty. Use at your own risk.
- The service is not responsible for any content accessed via the gateway.
- Please be aware that the SMS messages are not encrypted, so please do not use
  it for sensitive content. The provider and government may read your messages.
- The service is not affiliated with any organization. It is a personal project.
- This is not a commercial service or ISP. It is a free service provided by an
  individual. Subscribing is not meant to be a commercial transaction, but a
  way to support the service. Please do not abuse the service by sending spam
  or other malicious content.
- The service is not responsible for any content accessed via the gateway.
- Please be aware that the SMS messages are not encrypted, so please do not use
  it for sensitive content. The provider and government may read your messages.
- The service is not affiliated with any organization. It is a personal project.
- This is not a commercial service or ISP. It is a free service provided by an
  individual.`
		return c.SendString(doc)
	})

	// Add URI to Markdown endpoint
	app.Get("/uri2md", func(c *fiber.Ctx) error {
		uri := c.Query("uri")
		if uri == "" {
			return c.Status(400).SendString("Missing uri parameter")
		}

		markdown, err := fetchMarkdown(uri)
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Error fetching markdown: %v", err))
		}

		// Check if base64 encoding is requested
		if c.Query("b64") == "true" {
			encoded := base64.StdEncoding.EncodeToString([]byte(markdown))
			return c.SendString(encoded)
		}

		return c.SendString(markdown)
	})

	// Add Twitter user endpoint
	app.Get("/twitter", func(c *fiber.Ctx) error {
		user := c.Query("user")
		if user == "" {
			return c.Status(400).SendString("Missing user parameter")
		}

		tweets, err := fetchTweets(user, 25)
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Error fetching tweets: %v", err))
		}

		// Check if base64 encoding is requested
		if c.Query("b64") == "true" {
			encoded := base64.StdEncoding.EncodeToString([]byte(tweets))
			return c.SendString(encoded)
		}

		return c.SendString(tweets)
	})

	// Add webhook endpoint for Buy Me a Coffee subscriptions
	app.Post("/webhook/buymeacoffee", func(c *fiber.Ctx) error {
		var webhook BuyMeACoffeeWebhook
		if err := c.BodyParser(&webhook); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Handle different webhook types
		switch webhook.Type {
		case "recurring_donation.started":
			// Save subscription with active status
			if err := saveSubscription(
				webhook.Data.PspID,
				webhook.Data.SupporterEmail,
				"active",
				time.Unix(webhook.Data.CurrentPeriodEnd, 0),
			); err != nil {
				return c.Status(500).SendString(fmt.Sprintf("Error saving subscription: %v", err))
			}

		case "recurring_donation.updated":
			// Check if subscription is paused or canceled
			if webhook.Data.Paused == "true" || webhook.Data.Canceled == "true" {
				if err := updateSubscriptionStatus(webhook.Data.PspID, "inactive"); err != nil {
					return c.Status(500).SendString(fmt.Sprintf("Error updating subscription status: %v", err))
				}
			}

		case "recurring_donation.cancelled":
			// Update subscription status to cancelled
			if err := updateSubscriptionStatus(webhook.Data.PspID, "cancelled"); err != nil {
				return c.Status(500).SendString(fmt.Sprintf("Error updating subscription status: %v", err))
			}
		}

		return c.SendStatus(200)
	})

	// Add DuckDuckGo search endpoint
	app.Get("/ddg", func(c *fiber.Ctx) error {
		query := c.Query("q")
		if query == "" {
			return c.Status(400).SendString("Missing q parameter")
		}

		results, err := fetchDuckDuckGoResults(query)
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Error fetching search results: %v", err))
		}

		// Check if base64 encoding is requested
		if c.Query("b64") == "true" {
			encoded := base64.StdEncoding.EncodeToString([]byte(results))
			return c.SendString(encoded)
		}

		return c.SendString(results)
	})

	// Add Wikipedia summary endpoint
	app.Get("/wiki", func(c *fiber.Ctx) error {
		query := c.Query("q")
		if query == "" {
			return c.Status(400).SendString("Missing q parameter")
		}

		langCode := c.Query("lang")
		if langCode == "" {
			langCode = "en" // Default to English if no language specified
		}

		summary, err := fetchWikipediaSummary(query, langCode)
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Error fetching Wikipedia summary: %v", err))
		}

		// Check if base64 encoding is requested
		if c.Query("b64") == "true" {
			encoded := base64.StdEncoding.EncodeToString([]byte(summary))
			return c.SendString(encoded)
		}

		return c.SendString(summary)
	})

	// Add Weather forecast endpoint
	app.Get("/weather", func(c *fiber.Ctx) error {
		location := c.Query("loc")
		if location == "" {
			return c.Status(400).SendString("Missing loc parameter")
		}

		forecast, err := fetchWeatherForecast(location)
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Error fetching weather forecast: %v", err))
		}

		// Weather is sent as is, without base64 encoding
		return c.SendString(forecast)
	})

	// Add test endpoint that echoes back the SMS payload
	app.Post("/api/test", func(c *fiber.Ctx) error {
		// Deny access in production
		if currentEnv == EnvProd {
			return c.Status(403).SendString("Test endpoint is not available in production environment")
		}

		var payload []SMSPayload
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Log received payload
		receivedJSON, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Printf("Test endpoint received payload:\n%s\n", string(receivedJSON))

		// Process the payload and create response
		var response []providers.Message
		for _, sms := range payload {
			content := strings.TrimSpace(sms.Content)

			// Check if content is a URL
			if isURL(content) {
				markdown, err := fetchMarkdown(content)
				if err != nil {
					fmt.Println("Error fetching markdown:", err)
					continue
				}

				// Split and encode the message
				encodedParts := splitAndEncodeMessage(markdown, 500)

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
				tweets, err := fetchTweets(username, 5)
				if err != nil {
					fmt.Println("Error fetching tweets:", err)
					continue
				}

				// Split and encode the message
				encodedParts := splitAndEncodeMessage(tweets, 500)

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
				results, err := fetchDuckDuckGoResults(query)
				if err != nil {
					fmt.Println("Error fetching search results:", err)
					continue
				}

				// Split and encode the message
				encodedParts := splitAndEncodeMessage(results, 500)

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

				summary, err := fetchWikipediaSummary(query, langCode)
				if err != nil {
					fmt.Println("Error fetching Wikipedia summary:", err)
					continue
				}

				// Split and encode the message
				encodedParts := splitAndEncodeMessage(summary, 500)

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
				forecast, err := fetchWeatherForecast(location)
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
		responsePayload := SMSRequest{
			Username:   os.Getenv("SMS_USERNAME"),
			Password:   os.Getenv("SMS_PASSWORD"),
			SourceAddr: os.Getenv("SMS_SOURCE_ADDR"),
			ValidFor:   "48:00",
			Datacoding: "0",
			Messages:   response,
		}

		return c.JSON(responsePayload)
	})

	app.Post("/api/inbound", func(c *fiber.Ctx) error {
		var payload []SMSPayload
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Log received payload
		receivedJSON, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Printf("Received payload:\n%s\n", string(receivedJSON))

		for _, sms := range payload {
			content := strings.TrimSpace(sms.Content)

			// Handle subscription command
			if strings.HasPrefix(strings.ToLower(content), "subscribe ") {
				email := strings.TrimSpace(strings.TrimPrefix(content, "subscribe"))
				if err := linkPhoneToSubscription(sms.SourceAddr, email); err != nil {
					fmt.Printf("Error linking phone to subscription: %v\n", err)
					continue
				}
				// Update rate limit for the phone number
				if err := updateRateLimitForPhone(sms.SourceAddr, 20); err != nil {
					fmt.Printf("Error updating rate limit: %v\n", err)
					continue
				}
				// Send confirmation message
				confirmationMsg := "Your subscription has been activated. You now have a limit of 20 messages per hour."
				encodedParts := splitAndEncodeMessage(confirmationMsg, 500)
				for i, encoded := range encodedParts {
					smsMessage := []providers.Message{
						{
							Msg:  encoded,
							Dest: sms.SourceAddr,
							ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), i),
						},
					}
					if err := sendSMS(smsMessage); err != nil {
						fmt.Printf("Error sending confirmation message: %v\n", err)
					}
				}
				continue
			}

			// Check rate limit
			allowed, err := checkRateLimit(sms.SourceAddr)
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
				if err := sendSMS(smsMessage); err != nil {
					fmt.Printf("Error sending rate limit notification: %v\n", err)
				}
				continue
			}

			// Check if content is a URL
			if isURL(content) {
				markdown, err := fetchMarkdown(content)
				if err != nil {
					fmt.Println("Error fetching markdown:", err)
					continue
				}

				// Split and encode the message
				encodedParts := splitAndEncodeMessage(markdown, 500)

				// Send SMS with markdown content
				for i, encoded := range encodedParts {
					smsMessage := []providers.Message{
						{
							Msg:  encoded,
							Dest: sms.SourceAddr,
							ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), i),
						},
					}
					if err := sendSMS(smsMessage); err != nil {
						fmt.Println("Error sending SMS:", err)
					}
				}
				continue
			}

			// Check if content matches "twitter user <username>"
			if strings.HasPrefix(strings.ToLower(content), "twitter user ") {
				username := strings.TrimSpace(strings.TrimPrefix(content, "twitter user"))
				tweets, err := fetchTweets(username, 5)
				if err != nil {
					fmt.Println("Error fetching tweets:", err)
					continue
				}

				// Split and encode the message
				encodedParts := splitAndEncodeMessage(tweets, 500)

				// Send SMS with tweets content
				for i, encoded := range encodedParts {
					smsMessage := []providers.Message{
						{
							Msg:  encoded,
							Dest: sms.SourceAddr,
							ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), i),
						},
					}
					if err := sendSMS(smsMessage); err != nil {
						fmt.Println("Error sending SMS:", err)
					}
				}
				continue
			}

			// Check if content matches "websearch <query>"
			if strings.HasPrefix(strings.ToLower(content), "websearch ") {
				query := strings.TrimSpace(strings.TrimPrefix(content, "websearch"))
				results, err := fetchDuckDuckGoResults(query)
				if err != nil {
					fmt.Println("Error fetching search results:", err)
					continue
				}

				// Split and encode the message
				encodedParts := splitAndEncodeMessage(results, 500)

				// Send SMS with search results
				for i, encoded := range encodedParts {
					smsMessage := []providers.Message{
						{
							Msg:  encoded,
							Dest: sms.SourceAddr,
							ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), i),
						},
					}
					if err := sendSMS(smsMessage); err != nil {
						fmt.Println("Error sending SMS:", err)
					}
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

				summary, err := fetchWikipediaSummary(query, langCode)
				if err != nil {
					fmt.Println("Error fetching Wikipedia summary:", err)
					continue
				}

				// Split and encode the message
				encodedParts := splitAndEncodeMessage(summary, 500)

				// Send SMS with Wikipedia summary
				for i, encoded := range encodedParts {
					smsMessage := []providers.Message{
						{
							Msg:  encoded,
							Dest: sms.SourceAddr,
							ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), i),
						},
					}
					if err := sendSMS(smsMessage); err != nil {
						fmt.Println("Error sending SMS:", err)
					}
				}
				continue
			}

			// Check if content matches "weather <location>"
			if strings.HasPrefix(strings.ToLower(content), "weather ") {
				location := strings.TrimSpace(strings.TrimPrefix(content, "weather"))
				forecast, err := fetchWeatherForecast(location)
				if err != nil {
					fmt.Println("Error fetching weather forecast:", err)
					continue
				}

				// Weather is sent without encoding
				smsMessage := []providers.Message{
					{
						Msg:  forecast,
						Dest: sms.SourceAddr,
						ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), 0),
					},
				}
				if err := sendSMS(smsMessage); err != nil {
					fmt.Println("Error sending SMS:", err)
				}
				continue
			}
		}
		return c.SendStatus(204)
	})

	// Add test endpoint for subscription
	app.Post("/api/test/subscribe", func(c *fiber.Ctx) error {
		// Deny access in production
		if currentEnv == EnvProd {
			return c.Status(403).SendString("Test endpoint is not available in production environment")
		}

		type SubscriptionRequest struct {
			Email string `json:"email"`
		}

		var req SubscriptionRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		if req.Email == "" {
			return c.Status(400).SendString("Email is required")
		}

		// Save subscription with active status
		if err := saveSubscription(
			fmt.Sprintf("test_%d", time.Now().Unix()),
			req.Email,
			"active",
			time.Now().Add(30*24*time.Hour), // 30 days subscription
		); err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Error saving subscription: %v", err))
		}

		return c.SendString("Subscription added successfully")
	})

	app.Listen(":8080")
}

func isURL(text string) bool {
	re := regexp.MustCompile(`https?://[\w\-\.]+[\w\-/]*`)
	return re.MatchString(text)
}

func fetchMarkdown(url string) (string, error) {
	resp, err := httpClient.Get(fmt.Sprintf("https://urltomarkdown.herokuapp.com/?clean=true&url=%s", url))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	markdown, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(markdown), nil
}

func fetchTweets(username string, count int) (string, error) {
	nitterURL := fmt.Sprintf("https://nitter.app.ooguz.dev/%s/rss", username)
	resp, err := httpClient.Get(nitterURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tweets, err := parseTweetsFromRSS(string(body), count)
	if err != nil {
		return "", err
	}

	return tweets, nil
}

func parseTweetsFromRSS(rss string, count int) (string, error) {
	var feed RSS
	if err := xml.Unmarshal([]byte(rss), &feed); err != nil {
		return "", fmt.Errorf("error parsing RSS: %v", err)
	}

	if len(feed.Channel.Items) == 0 {
		return "", fmt.Errorf("no tweets found")
	}

	var tweets []string
	// Process items until we have enough non-retweet tweets
	for i := 0; i < len(feed.Channel.Items) && len(tweets) < count; i++ {
		item := feed.Channel.Items[i]

		// Clean up HTML entities
		title := item.Title
		title = strings.ReplaceAll(title, "&quot;", "\"")
		title = strings.ReplaceAll(title, "&apos;", "'")
		title = strings.ReplaceAll(title, "&lt;", "<")
		title = strings.ReplaceAll(title, "&gt;", ">")
		title = strings.ReplaceAll(title, "&amp;", "&")

		// Skip RT by @username: tweets
		if strings.HasPrefix(title, "RT by @") {
			continue
		}

		// Skip empty titles
		if strings.TrimSpace(title) == "" {
			continue
		}

		// Add the tweet
		tweets = append(tweets, fmt.Sprintf("- %s", title))
	}

	if len(tweets) == 0 {
		return "", fmt.Errorf("no tweets found")
	}

	return strings.Join(tweets, "\n"), nil
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func splitMessage(message string, maxLength int) []string {
	if len(message) <= maxLength {
		return []string{message}
	}

	var parts []string
	currentPart := ""

	// Split by newlines to keep content together
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		if len(currentPart)+len(line)+1 > maxLength { // +1 for newline
			if currentPart != "" {
				parts = append(parts, currentPart)
			}
			currentPart = line
		} else {
			if currentPart != "" {
				currentPart += "\n"
			}
			currentPart += line
		}
	}

	if currentPart != "" {
		parts = append(parts, currentPart)
	}

	return parts
}

func splitAndEncodeMessage(message string, maxLength int) []string {
	// Split the message into parts
	parts := splitMessage(message, maxLength)

	// Encode each part with a header
	var encodedParts []string
	for i, part := range parts {
		// First encode the message part
		encoded := base64.StdEncoding.EncodeToString([]byte(part))
		// Then add the header to the encoded content
		header := fmt.Sprintf("GW%d|", i+1)
		encodedParts = append(encodedParts, header+encoded)
	}

	return encodedParts
}

func sendSMS(messages []providers.Message) error {
	return smsManager.SendMessage(messages)
}
