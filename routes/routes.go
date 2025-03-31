package routes

import (
	"smsgw/controllers"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(
	app *fiber.App,
	docController *controllers.DocController,
	contentController *controllers.ContentController,
	webhookController *controllers.WebhookController,
	smsController *controllers.SMSController,
) {
	// Documentation routes
	app.Get("/", docController.HandleRoot)
	app.Get("/index.html", docController.HandleIndexHTML)
	app.Get("/index.htm", docController.HandleIndexHTML)
	app.Get("/index.txt", docController.HandleIndexText)
	app.Get("/api/docs", docController.HandleAPIDocs)
	app.Get("/api/docs/ui", docController.HandleAPIDocsUI)

	// Content routes
	app.Get("/uri2md", contentController.HandleURI2MD)
	app.Get("/twitter", contentController.HandleTwitter)
	app.Get("/ddg", contentController.HandleDuckDuckGo)
	app.Get("/wiki", contentController.HandleWikipedia)
	app.Get("/weather", contentController.HandleWeather)

	// Webhook routes
	app.Post("/webhook/buymeacoffee", webhookController.HandleBuyMeACoffee)
	app.Post("/webhook/paypal", webhookController.HandlePayPal)

	// SMS routes
	app.Post("/api/inbound", smsController.HandleInbound)
	app.Post("/api/test", smsController.HandleTest)
	app.Post("/api/test/subscribe", smsController.HandleTestSubscribe)
}
