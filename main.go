package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

// Create HTTP client with timeout
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

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

type SMSMessage struct {
	Msg  string `json:"msg"`
	Dest string `json:"dest"`
	ID   string `json:"id"`
}

type SMSRequest struct {
	Username   string       `json:"username"`
	Password   string       `json:"password"`
	SourceAddr string       `json:"source_addr"`
	ValidFor   string       `json:"valid_for"`
	SendAt     string       `json:"send_at"`
	CustomID   string       `json:"custom_id"`
	Datacoding string       `json:"datacoding"`
	Messages   []SMSMessage `json:"messages"`
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
	}

	app := fiber.New()

	// Add root endpoint with documentation
	app.Get("/", func(c *fiber.Ctx) error {
		doc := `smsgw
=====

This SMS gateway provides a super minimal (and experimental!) HTTP-SMS gateway, can be considered as an emergency connection method similar to dial-up 

Usage:

It always response as base64. Send the URI (including the https://) as SMS and get markdown in base64

if you send "twitter user <username>" it will response with the last 5 tweets of the user


support: https://buymeacoffee.com/ooguz`
		return c.SendString(doc)
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

			// Check if content is a URL
			if isURL(content) {
				markdown, err := fetchMarkdown(content)
				if err != nil {
					fmt.Println("Error fetching markdown:", err)
					continue
				}
				encoded := base64.StdEncoding.EncodeToString([]byte(markdown))

				// Send SMS with markdown content
				smsMessage := []SMSMessage{
					{
						Msg:  encoded,
						Dest: sms.SourceAddr,
						ID:   fmt.Sprintf("%d", time.Now().Unix()),
					},
				}
				err = sendSMS(smsMessage)
				if err != nil {
					fmt.Println("Error sending SMS:", err)
				}
				continue
			}

			// Check if content matches "twitter user <username>"
			if strings.HasPrefix(strings.ToLower(content), "twitter user ") {
				username := strings.TrimSpace(strings.TrimPrefix(content, "twitter user"))
				tweets, err := fetchTweets(username)
				if err != nil {
					fmt.Println("Error fetching tweets:", err)
					continue
				}
				encoded := base64.StdEncoding.EncodeToString([]byte(tweets))

				// Send SMS with tweets content
				smsMessage := []SMSMessage{
					{
						Msg:  encoded,
						Dest: sms.SourceAddr,
						ID:   fmt.Sprintf("%d", time.Now().Unix()),
					},
				}
				err = sendSMS(smsMessage)
				if err != nil {
					fmt.Println("Error sending SMS:", err)
				}
				continue
			}
		}
		return c.SendStatus(204)
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

	markdown, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(markdown), nil
}

func fetchTweets(username string) (string, error) {
	nitterURL := fmt.Sprintf("https://nitter.app.ooguz.dev/%s/rss", username)
	resp, err := httpClient.Get(nitterURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tweets, err := parseTweetsFromRSS(string(body))
	if err != nil {
		return "", err
	}

	return tweets, nil
}

func parseTweetsFromRSS(rss string) (string, error) {
	// Find all title tags with their content
	titleRegex := regexp.MustCompile(`<title>(?:<!\[CDATA\[)?(.*?)(?:\]\]>)?</title>`)
	matches := titleRegex.FindAllStringSubmatch(rss, -1)

	if len(matches) < 2 { // Skeip the first match as it's the channel title
		return "", fmt.Errorf("no tweets found")
	}

	var tweets []string
	// Process up to 5 items, skipping the first match (channel title)
	startIdx := 1
	endIdx := min(startIdx+5, len(matches))
	for i := startIdx; i < endIdx; i++ {
		if len(matches[i]) > 1 {
			// Clean up HTML entities
			title := matches[i][1]
			title = strings.ReplaceAll(title, "&quot;", "\"")
			title = strings.ReplaceAll(title, "&apos;", "'")
			title = strings.ReplaceAll(title, "&lt;", "<")
			title = strings.ReplaceAll(title, "&gt;", ">")
			title = strings.ReplaceAll(title, "&amp;", "&")

			tweets = append(tweets, fmt.Sprintf("- %s", title))
		}
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

func sendSMS(messages []SMSMessage) error {
	// Split long messages before encoding
	var allMessages []SMSMessage
	for _, msg := range messages {
		// Decode base64 message
		decoded, err := base64.StdEncoding.DecodeString(msg.Msg)
		if err != nil {
			return fmt.Errorf("error decoding base64 message: %v", err)
		}

		// Split the decoded message
		parts := splitMessage(string(decoded), 500)

		// Encode each part and create new messages
		for i, part := range parts {
			encoded := base64.StdEncoding.EncodeToString([]byte(part))
			allMessages = append(allMessages, SMSMessage{
				Msg:  encoded,
				Dest: msg.Dest,
				ID:   fmt.Sprintf("%s_%d", msg.ID, i),
			})
		}
	}

	smsRequest := SMSRequest{
		Username:   os.Getenv("SMS_USERNAME"),
		Password:   os.Getenv("SMS_PASSWORD"),
		SourceAddr: os.Getenv("SMS_SOURCE_ADDR"),
		ValidFor:   "48:00",
		Datacoding: "0",
		Messages:   allMessages,
	}

	jsonData, err := json.Marshal(smsRequest)
	if err != nil {
		return fmt.Errorf("error marshaling SMS request: %v", err)
	}

	// Log sent payload
	fmt.Printf("Sending SMS payload:\n%s\n", string(jsonData))

	resp, err := httpClient.Post(
		"https://sms.verimor.com.tr/v2/send.json",
		"application/json",
		strings.NewReader(string(jsonData)),
	)
	if err != nil {
		return fmt.Errorf("error sending SMS: %v", err)
	}
	defer resp.Body.Close()

	// Log response
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("SMS API Response (Status: %d):\n%s\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SMS API error: %s", string(body))
	}

	return nil
}
