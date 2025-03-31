package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"smsgw/config"
	"smsgw/controllers"
	"smsgw/database"
	"smsgw/providers"
	"smsgw/routes"
	"smsgw/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

//go:embed static/*
var embedDirStatic embed.FS

func main() {

	// Initialize config
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Start periodic privacy-focused cleanup routine
	go startPrivacyCleanupJob(db)

	// Initialize SMS provider manager
	smsManager := providers.NewManager()
	smsManager.RegisterProvider(providers.NewVerimorProvider())

	// Initialize services
	markdownService := services.NewMarkdownService(cfg.HTTPClient)
	twitterService := services.NewTwitterService(cfg.HTTPClient)
	searchService := services.NewSearchService(cfg.HTTPClient)
	weatherService := services.NewWeatherService(cfg.HTTPClient)
	subscriptionService := services.NewSubscriptionService()
	smsService := services.NewSMSService(smsManager)

	// Initialize controllers
	docController := controllers.NewDocController(cfg)
	contentController := controllers.NewContentController(
		markdownService,
		twitterService,
		searchService,
		weatherService,
	)
	webhookController := controllers.NewWebhookController(subscriptionService)
	smsController := controllers.NewSMSController(
		&controllers.Config{Environment: string(cfg.Environment)},
		smsService,
		markdownService,
		twitterService,
		searchService,
		weatherService,
		subscriptionService,
	)

	// Initialize Fiber app
	app := fiber.New(
		fiber.Config{
			AppName:      "neo146 - infogate",
			Prefork:      true,
			ServerHeader: "neo146-infogate",
		},
	)
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(requestid.New())
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
	}))
	app.Use("/static", filesystem.New(filesystem.Config{
		Root:       http.FS(embedDirStatic),
		PathPrefix: "static",
		Browse:     false,
	}))
	app.Use(favicon.New(favicon.Config{
		FileSystem: http.FS(embedDirStatic),
		File:       "static/favicon.ico",
	}))

	// Set up routes
	routes.SetupRoutes(app, docController, contentController, webhookController, smsController)

	// Start server
	log.Fatal(app.Listen(":8080"))
}

// startPrivacyCleanupJob runs periodic cleanup of rate limit data to ensure user privacy
func startPrivacyCleanupJob(db *database.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if err := db.PurgeOldMessageData(); err != nil {
			log.Printf("Error purging old message data: %v", err)
		}
	}
}
