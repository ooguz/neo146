package controllers

import (
	"encoding/base64"
	"smsgw/services"

	"github.com/gofiber/fiber/v2"
)

// ContentController handles content-related endpoints
type ContentController struct {
	markdownService *services.MarkdownService
	twitterService  *services.TwitterService
	searchService   *services.SearchService
	weatherService  *services.WeatherService
}

// NewContentController creates a new ContentController
func NewContentController(
	markdownService *services.MarkdownService,
	twitterService *services.TwitterService,
	searchService *services.SearchService,
	weatherService *services.WeatherService,
) *ContentController {
	return &ContentController{
		markdownService: markdownService,
		twitterService:  twitterService,
		searchService:   searchService,
		weatherService:  weatherService,
	}
}

// HandleURI2MD handles the URI to Markdown endpoint
func (c *ContentController) HandleURI2MD(ctx *fiber.Ctx) error {
	uri := ctx.Query("uri")
	if uri == "" {
		return ctx.Status(400).SendString("Missing uri parameter")
	}

	markdown, err := c.markdownService.FetchMarkdown(uri)
	if err != nil {
		return ctx.Status(500).SendString("Error fetching markdown: " + err.Error())
	}

	// Check if base64 encoding is requested
	if ctx.Query("b64") == "true" {
		encoded := base64.StdEncoding.EncodeToString([]byte(markdown))
		return ctx.SendString(encoded)
	}

	return ctx.SendString(markdown)
}

// HandleTwitter handles the Twitter user endpoint
func (c *ContentController) HandleTwitter(ctx *fiber.Ctx) error {
	user := ctx.Query("user")
	if user == "" {
		return ctx.Status(400).SendString("Missing user parameter")
	}

	tweets, err := c.twitterService.FetchTweets(user, 25)
	if err != nil {
		return ctx.Status(500).SendString("Error fetching tweets: " + err.Error())
	}

	// Check if base64 encoding is requested
	if ctx.Query("b64") == "true" {
		encoded := base64.StdEncoding.EncodeToString([]byte(tweets))
		return ctx.SendString(encoded)
	}

	return ctx.SendString(tweets)
}

// HandleDuckDuckGo handles the DuckDuckGo search endpoint
func (c *ContentController) HandleDuckDuckGo(ctx *fiber.Ctx) error {
	query := ctx.Query("q")
	if query == "" {
		return ctx.Status(400).SendString("Missing q parameter")
	}

	results, err := c.searchService.FetchDuckDuckGoResults(query)
	if err != nil {
		return ctx.Status(500).SendString("Error fetching search results: " + err.Error())
	}

	// Check if base64 encoding is requested
	if ctx.Query("b64") == "true" {
		encoded := base64.StdEncoding.EncodeToString([]byte(results))
		return ctx.SendString(encoded)
	}

	return ctx.SendString(results)
}

// HandleWikipedia handles the Wikipedia summary endpoint
func (c *ContentController) HandleWikipedia(ctx *fiber.Ctx) error {
	query := ctx.Query("q")
	if query == "" {
		return ctx.Status(400).SendString("Missing q parameter")
	}

	langCode := ctx.Query("lang")
	if langCode == "" {
		langCode = "en" // Default to English if no language specified
	}

	summary, err := c.searchService.FetchWikipediaSummary(query, langCode)
	if err != nil {
		return ctx.Status(500).SendString("Error fetching Wikipedia summary: " + err.Error())
	}

	// Check if base64 encoding is requested
	if ctx.Query("b64") == "true" {
		encoded := base64.StdEncoding.EncodeToString([]byte(summary))
		return ctx.SendString(encoded)
	}

	return ctx.SendString(summary)
}

// HandleWeather handles the weather forecast endpoint
func (c *ContentController) HandleWeather(ctx *fiber.Ctx) error {
	location := ctx.Query("loc")
	if location == "" {
		return ctx.Status(400).SendString("Missing loc parameter")
	}

	forecast, err := c.weatherService.FetchWeatherForecast(location)
	if err != nil {
		return ctx.Status(500).SendString("Error fetching weather forecast: " + err.Error())
	}

	// Weather is sent as is, without base64 encoding
	return ctx.SendString(forecast)
}
