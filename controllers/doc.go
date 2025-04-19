package controllers

import (
	"embed"
	"neo146/config"
	"strings"

	"github.com/gofiber/fiber/v2"
)

//go:embed static/*
var staticFS embed.FS

// DocController handles documentation-related endpoints
type DocController struct {
	config        *config.Config
	rootInfoText  string
	rootInfoHTML  string
	swaggerUIHTML string
}

// NewDocController creates a new DocController
func NewDocController(config *config.Config) *DocController {
	// Load the static files
	rootInfoTextBytes, _ := staticFS.ReadFile("static/root_info.txt")
	rootInfoHTMLBytes, _ := staticFS.ReadFile("static/index.html")
	swaggerUIHTMLBytes, _ := staticFS.ReadFile("static/swagger.html")

	return &DocController{
		config:        config,
		rootInfoText:  string(rootInfoTextBytes),
		rootInfoHTML:  string(rootInfoHTMLBytes),
		swaggerUIHTML: string(swaggerUIHTMLBytes),
	}
}

// HandleRoot handles the root endpoint
func (c *DocController) HandleRoot(ctx *fiber.Ctx) error {
	// Get user agent
	userAgent := ctx.Get("User-Agent")

	// Check if user agent is a modern web browser
	isBrowser := strings.Contains(strings.ToLower(userAgent), "firefox") ||
		strings.Contains(strings.ToLower(userAgent), "chrome") ||
		strings.Contains(strings.ToLower(userAgent), "safari") ||
		strings.Contains(strings.ToLower(userAgent), "edge") ||
		strings.Contains(strings.ToLower(userAgent), "opera")

	if isBrowser {
		ctx.Set("Content-Type", "text/html")
		return ctx.Status(200).SendString(c.rootInfoHTML)
	}

	// Default to text for non-browser user agents
	return ctx.SendString(c.rootInfoText)
}

// HandleIndexHTML serves the HTML version
func (c *DocController) HandleIndexHTML(ctx *fiber.Ctx) error {
	ctx.Set("Content-Type", "text/html")
	return ctx.Status(200).SendString(c.rootInfoHTML)
}

// HandleIndexText serves the text version
func (c *DocController) HandleIndexText(ctx *fiber.Ctx) error {
	return ctx.SendString(c.rootInfoText)
}

// HandleAPIDocs handles the API documentation endpoint
func (c *DocController) HandleAPIDocs(ctx *fiber.Ctx) error {
	return ctx.Status(200).Type("application/json").SendString(c.config.OpenAPISpec)
}

// HandleAPIDocsUI handles the API documentation UI endpoint
func (c *DocController) HandleAPIDocsUI(ctx *fiber.Ctx) error {
	ctx.Set("Content-Type", "text/html")
	return ctx.Status(200).SendString(c.swaggerUIHTML)
}
