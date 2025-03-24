# SMS Gateway

This is a simple HTTP-SMS gateway that provides a minimal "dial-up" style experience through SMS.

## Features

* **URL to Markdown**: Send any URL as SMS and get the content as markdown
* **Twitter/Nitter**: Get the latest tweets from a user
* **DuckDuckGo Search**: Search the web and get results via SMS
* **Wikipedia**: Get article summaries from Wikipedia in different languages
* **Weather**: Get weather forecasts for any location using wttr.in

## Usage

### SMS Commands

* `https://example.com` - Get a markdown version of the URL
* `twitter user <username>` - Get the last 5 tweets from a user
* `websearch <query>` - Search the web using DuckDuckGo
* `wiki <2charlangcode> <query>` - Get Wikipedia article summary (e.g., `wiki en bitcoin`)
* `weather <location>` - Get weather forecast (e.g., `weather istanbul`)

### HTTP Endpoints

* `/uri2md?uri=<uri>[&b64=true]` - Convert URI to Markdown
* `/twitter?user=<user>[&b64=true]` - Get last 5 tweets of a user
* `/ddg?q=<query>[&b64=true]` - Search the web via DuckDuckGo
* `/wiki?lang=<2charlangcode>&q=<query>[&b64=true]` - Get Wikipedia article summary
* `/weather?loc=<location>` - Get weather forecast (sent without base64 encoding)

Adding `b64=true` will return the response base64 encoded (not applicable to weather endpoint).

### Weather Location Formats

The weather feature supports various location formats:
* City name (e.g., `istanbul`)
* Any location (e.g., `~Saraçhane`, use `+` for spaces)
* Unicode location names in any language (e.g., `Москва`)
* Airport codes (3 letters, e.g., `esb`)
* Domain names (e.g., `@stackoverflow.com`)
* Postal codes (e.g., `06800`)
* GPS coordinates (e.g., `39.925325,32.836987`)

## Rate Limits

* Free users: 5 messages per hour per phone number
* Subscribed users: 20 messages per hour per phone number
* Subscription link: https://buymeacoffee.com/ooguz

## Environment Variables

Required environment variables:
* `SMS_USERNAME`: Username for SMS provider
* `SMS_PASSWORD`: Password for SMS provider
* `SMS_SOURCE_ADDR`: Source address for SMS
* `SMS_PROVIDER`: Provider name (default: "Verimor")
* `APP_ENV`: Application environment ("test" or "prod")

## Adding SMS Providers

The system is designed to support multiple SMS providers. To add a new provider:

1. Create a new file in the `providers` directory
2. Implement the `Provider` interface
3. Register the provider in `main.go`

## License

This project is open source. 