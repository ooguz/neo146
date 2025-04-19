package services

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"neo146/utils"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	botMutex    sync.Mutex
	botInstance *TelegramService
)

// TelegramService handles Telegram bot operations
type TelegramService struct {
	bot         *tgbotapi.BotAPI
	smsService  *SMSService
	subService  *SubscriptionService
	rateLimits  map[int64]int
	lastMessage map[int64]time.Time
}

// NewTelegramService creates a new Telegram service
func NewTelegramService(token string, smsService *SMSService, subService *SubscriptionService) (*TelegramService, error) {
	botMutex.Lock()
	defer botMutex.Unlock()

	if botInstance != nil {
		return nil, fmt.Errorf("bot instance already exists")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("error creating Telegram bot: %v", err)
	}

	botInstance = &TelegramService{
		bot:         bot,
		smsService:  smsService,
		subService:  subService,
		rateLimits:  make(map[int64]int),
		lastMessage: make(map[int64]time.Time),
	}

	return botInstance, nil
}

// Start starts the Telegram bot
func (t *TelegramService) Start() error {
	// Try to acquire lock
	if err := utils.AcquireLock(); err != nil {
		// If we get a "bot is already running" error, try to force cleanup
		if err.Error() == "bot is already running" {
			log.Println("Force cleaning up stale lock file...")
			utils.ReleaseLock()
			// Try to acquire lock again
			if err := utils.AcquireLock(); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer utils.ReleaseLock()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Handle commands
		if update.Message.IsCommand() {
			t.handleCommand(update.Message)
			continue
		}

		// Handle regular messages
		t.handleMessage(update.Message)
	}

	return nil
}

// Cleanup cleans up the bot instance
func (t *TelegramService) Cleanup() {
	botMutex.Lock()
	defer botMutex.Unlock()

	if botInstance == t {
		botInstance = nil
	}
	// Ensure lock file is released
	utils.ReleaseLock()
}

// handleCommand processes Telegram bot commands
func (t *TelegramService) handleCommand(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	command := message.Command()

	switch command {
	case "start":
		t.sendMessage(chatID, `neo146 provides a minimal (and experimental!) information gateway that serves as an emergency network connection method inspired by dial-up, allowing you to access content via certain protocols. The current implementations are HTTP-SMS gateway, HTTP-Markdown gateway and Telegram gateway.

Send /help for available options.

Running this service costs about 20 EUR per month. For a better experience and support the service, please consider subscribing.

https://buymeacoffee.com/ooguz`)
	case "help":
		t.sendMessage(chatID, `Available commands:
/url <url> - Convert webpage to Markdown
/twitter <username> - Get last 5 tweets
/search <query> - Search the web
/wiki <lang> <query> - Get Wikipedia summary
/weather <location> - Get weather forecast
/subscribe <email> - Subscribe to the service`)
	case "url":
		url := message.CommandArguments()
		if url == "" {
			t.sendMessage(chatID, "Please provide a URL. Usage: /url <url>")
			return
		}
		t.handleURL(chatID, url)
	case "twitter":
		username := message.CommandArguments()
		if username == "" {
			t.sendMessage(chatID, "Please provide a username. Usage: /twitter <username>")
			return
		}
		t.handleTwitter(chatID, username)
	case "search":
		query := message.CommandArguments()
		if query == "" {
			t.sendMessage(chatID, "Please provide a search query. Usage: /search <query>")
			return
		}
		t.handleSearch(chatID, query)
	case "wiki":
		args := strings.SplitN(message.CommandArguments(), " ", 2)
		if len(args) != 2 {
			t.sendMessage(chatID, "Please provide language code and query. Usage: /wiki <lang> <query>")
			return
		}
		t.handleWiki(chatID, args[0], args[1])
	case "weather":
		location := message.CommandArguments()
		if location == "" {
			t.sendMessage(chatID, "Please provide a location. Usage: /weather <location>")
			return
		}
		t.handleWeather(chatID, location)
	case "subscribe":
		email := message.CommandArguments()
		if email == "" {
			t.sendMessage(chatID, "Please provide your email. Usage: /subscribe <email>")
			return
		}
		t.handleSubscribe(chatID, email)
	default:
		t.sendMessage(chatID, "Unknown command. Use /help to see available commands.")
	}
}

// handleMessage processes regular messages
func (t *TelegramService) handleMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	text := message.Text

	// Check if the message is a URL
	if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") {
		t.handleURL(chatID, text)
		return
	}

	t.sendMessage(chatID, "Please use commands to interact with the bot. Use /help to see available commands.")
}

// handleURL processes URL requests
func (t *TelegramService) handleURL(chatID int64, url string) {
	// Check rate limit
	if !t.checkRateLimit(chatID) {
		t.sendMessage(chatID, "Rate limit exceeded. Please try again later.")
		return
	}

	// Get full markdown content
	markdown, err := t.smsService.MarkdownService.FetchMarkdown(url)
	if err != nil {
		t.sendMessage(chatID, fmt.Sprintf("Error fetching markdown: %v", err))
		return
	}

	// Sanitize and encode the content
	sanitizedContent := sanitizeContent(markdown)

	// Send content in chunks to avoid message length limits
	chunks := splitIntoChunks(sanitizedContent, 4000)
	for _, chunk := range chunks {
		t.sendMessage(chatID, chunk)
		time.Sleep(100 * time.Millisecond) // Small delay between messages
	}
}

// sanitizeContent ensures the content is UTF-8 encoded and removes problematic characters
func sanitizeContent(content string) string {
	// Convert to UTF-8 if needed
	content = string([]rune(content))

	// Remove any problematic characters
	content = strings.Map(func(r rune) rune {
		// Remove control characters
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		// Remove other problematic characters
		if r == '\u200B' || r == '\u200C' || r == '\u200D' || r == '\uFEFF' {
			return -1
		}
		return r
	}, content)

	// Replace any remaining invalid UTF-8 sequences
	content = strings.ToValidUTF8(content, "")

	// Ensure the content is not empty after sanitization
	if content == "" {
		content = "Message content was empty after sanitization"
	}

	return content
}

// handleTwitter processes Twitter requests
func (t *TelegramService) handleTwitter(chatID int64, username string) {
	// Check rate limit
	if !t.checkRateLimit(chatID) {
		t.sendMessage(chatID, "Rate limit exceeded. Please try again later.")
		return
	}

	// Get more tweets for Telegram (10 instead of 5)
	tweets, err := t.smsService.TwitterService.FetchTweets(username, 10)
	if err != nil {
		t.sendMessage(chatID, fmt.Sprintf("Error fetching tweets: %v", err))
		return
	}

	// Send tweets in chunks to avoid message length limits
	chunks := splitIntoChunks(tweets, 4000)
	for _, chunk := range chunks {
		t.sendMessage(chatID, chunk)
		time.Sleep(100 * time.Millisecond) // Small delay between messages
	}
}

// handleSearch processes web search requests
func (t *TelegramService) handleSearch(chatID int64, query string) {
	// Check rate limit
	if !t.checkRateLimit(chatID) {
		t.sendMessage(chatID, "Rate limit exceeded. Please try again later.")
		return
	}

	// Get more detailed search results
	results, err := t.smsService.SearchService.FetchDuckDuckGoResults(query)
	if err != nil {
		t.sendMessage(chatID, fmt.Sprintf("Error fetching search results: %v", err))
		return
	}

	// Send results in chunks to avoid message length limits
	chunks := splitIntoChunks(results, 4000)
	for _, chunk := range chunks {
		t.sendMessage(chatID, chunk)
		time.Sleep(100 * time.Millisecond) // Small delay between messages
	}
}

// fetchDetailedWikiContent fetches comprehensive Wikipedia content
func (t *TelegramService) fetchDetailedWikiContent(langCode, query string) (string, error) {
	// Base URL for Wikipedia REST API
	baseURL := fmt.Sprintf("https://%s.wikipedia.org/api/rest_v1", langCode)

	// Fetch page content
	contentURL := fmt.Sprintf("%s/page/html/%s", baseURL, query)
	resp, err := http.Get(contentURL)
	if err != nil {
		return "", fmt.Errorf("error fetching Wikipedia content: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Wikipedia API returned status %d", resp.StatusCode)
	}

	// Parse HTML content
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error parsing Wikipedia content: %v", err)
	}

	// Extract main content
	var content strings.Builder
	content.WriteString(fmt.Sprintf("*Wikipedia Article: %s*\n\n", query))

	// Track processed sections to avoid duplicates
	processedSections := make(map[string]bool)

	// Get the main content
	doc.Find("section").Each(func(i int, s *goquery.Selection) {
		// Skip unwanted sections
		if s.HasClass("references") || s.HasClass("notes") ||
			s.HasClass("see_also") || s.HasClass("bibliography") ||
			s.HasClass("external_links") {
			return
		}

		// Get section title
		title := s.Find("h1, h2, h3, h4, h5, h6").Text()
		if title != "" {
			// Skip if we've already processed this section
			if processedSections[title] {
				return
			}
			processedSections[title] = true
			content.WriteString(fmt.Sprintf("*%s*\n", title))
		}

		// Get section content
		s.Find("p").Each(func(i int, p *goquery.Selection) {
			// Remove reference numbers and citations
			p.Find("sup").Remove()
			text := strings.TrimSpace(p.Text())
			if text != "" {
				content.WriteString(fmt.Sprintf("%s\n\n", text))
			}
		})
	})

	// Add article link
	content.WriteString(fmt.Sprintf("\n[Read more on Wikipedia](https://%s.wikipedia.org/wiki/%s)", langCode, query))

	return content.String(), nil
}

// handleWiki processes Wikipedia requests
func (t *TelegramService) handleWiki(chatID int64, langCode, query string) {
	// Check rate limit
	if !t.checkRateLimit(chatID) {
		t.sendMessage(chatID, "Rate limit exceeded. Please try again later.")
		return
	}

	// Get detailed Wikipedia content
	content, err := t.fetchDetailedWikiContent(langCode, query)
	if err != nil {
		// Fallback to simple summary if detailed content fails
		summary, err := t.smsService.SearchService.FetchWikipediaSummary(query, langCode)
		if err != nil {
			t.sendMessage(chatID, fmt.Sprintf("Error fetching Wikipedia content: %v", err))
			return
		}
		content = fmt.Sprintf("*Wikipedia Article: %s*\n\n%s", query, summary)
	}

	// Send content in chunks to avoid message length limits
	chunks := splitIntoChunks(content, 4000)
	for i, chunk := range chunks {
		// Add page indicator for multi-part messages
		if len(chunks) > 1 {
			chunk = fmt.Sprintf("Page %d/%d\n\n%s", i+1, len(chunks), chunk)
		}
		t.sendMessage(chatID, chunk)
		time.Sleep(100 * time.Millisecond) // Small delay between messages
	}
}

// handleWeather processes weather requests
func (t *TelegramService) handleWeather(chatID int64, location string) {
	// Check rate limit
	if !t.checkRateLimit(chatID) {
		t.sendMessage(chatID, "Rate limit exceeded. Please try again later.")
		return
	}

	// Get detailed weather forecast
	forecast, err := t.smsService.WeatherService.FetchWeatherForecast(location)
	if err != nil {
		t.sendMessage(chatID, fmt.Sprintf("Error fetching weather forecast: %v", err))
		return
	}

	// Send forecast in chunks to avoid message length limits
	chunks := splitIntoChunks(forecast, 4000)
	for _, chunk := range chunks {
		t.sendMessage(chatID, chunk)
		time.Sleep(100 * time.Millisecond) // Small delay between messages
	}
}

// handleSubscribe processes subscription requests
func (t *TelegramService) handleSubscribe(chatID int64, email string) {
	// Generate a unique ID for the subscription
	subID := fmt.Sprintf("tg_%d_%d", chatID, time.Now().Unix())

	// Save subscription with active status
	if err := t.subService.SaveSubscription(
		subID,
		email,
		"active",
		time.Now().Add(30*24*time.Hour), // 30 days subscription
	); err != nil {
		t.sendMessage(chatID, fmt.Sprintf("Error saving subscription: %v", err))
		return
	}

	// Update rate limit for the chat
	t.rateLimits[chatID] = 20 // Subscribers get 20 messages per hour

	t.sendMessage(chatID, "Your subscription has been activated. You now have a limit of 20 messages per hour.")
}

// checkRateLimit checks if the user has exceeded their rate limit
func (t *TelegramService) checkRateLimit(chatID int64) bool {
	// Get current rate limit (default to 5 for non-subscribers)
	limit := t.rateLimits[chatID]
	if limit == 0 {
		limit = 5
	}

	// Check if enough time has passed since the last message
	lastMsgTime := t.lastMessage[chatID]
	if time.Since(lastMsgTime) < time.Hour {
		// If less than an hour has passed, check the count
		if t.rateLimits[chatID] <= 0 {
			return false
		}
		t.rateLimits[chatID]--
	} else {
		// Reset the rate limit counter
		t.rateLimits[chatID] = limit - 1
	}

	t.lastMessage[chatID] = time.Now()
	return true
}

// sendMessage sends a message to a Telegram chat
func (t *TelegramService) sendMessage(chatID int64, text string) {
	// Sanitize the message content
	text = sanitizeContent(text)

	// Create message
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	// Send message
	_, err := t.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending Telegram message: %v", err)
	}

	// Update last message time
	t.lastMessage[chatID] = time.Now()
}

// splitIntoChunks splits a string into chunks of specified size
func splitIntoChunks(text string, chunkSize int) []string {
	var chunks []string
	for i := 0; i < len(text); i += chunkSize {
		end := i + chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}
	return chunks
}
