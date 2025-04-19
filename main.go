package main

import (
	"embed"
	"fmt"
	"log"
	"neo146/config"
	"neo146/controllers"
	"neo146/database"
	"neo146/providers"
	"neo146/routes"
	"neo146/services"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	// Initialize HTTP client
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Initialize services
	providerManager := providers.NewManager()
	providerManager.RegisterProvider(providers.NewVerimorProvider())

	markdownService := services.NewMarkdownService(httpClient)
	twitterService := services.NewTwitterService(httpClient)
	searchService := services.NewSearchService(httpClient)
	weatherService := services.NewWeatherService(httpClient)
	subscriptionService := services.NewSubscriptionService()

	smsService := services.NewSMSService(providerManager)
	smsService.MarkdownService = markdownService
	smsService.TwitterService = twitterService
	smsService.SearchService = searchService
	smsService.WeatherService = weatherService

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

	// Initialize Telegram bot controller
	telegramController, err := controllers.NewTelegramController(smsService, subscriptionService)
	if err != nil {
		log.Printf("Error initializing Telegram bot: %v", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "neo146 - infogate",
		ServerHeader: "neo146-infogate",
	})

	// Middleware
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(recover.New())
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

	// Setup routes
	routes.SetupRoutes(app, docController, contentController, webhookController, smsController)

	// Start Telegram bot in a goroutine
	go func() {
		for {
			err := telegramController.StartBot()
			if err != nil {
				log.Printf("Failed to start Telegram bot: %v", err)
				time.Sleep(5 * time.Second) // Wait before retrying
				continue
			}
			break
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")
		telegramController.Cleanup()
		app.Shutdown()
	}()

	// Start server
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
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
