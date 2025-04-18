![neo146 logo](https://github.com/user-attachments/assets/56a5d90b-dbea-4637-8611-16aa17d61bae)

# neo146 - connect like it's 1984

![Contributors](https://img.shields.io/github/contributors/ooguz/neo146?color=dark-green) ![Stargazers](https://img.shields.io/github/stars/ooguz/neo146?style=social) ![Issues](https://img.shields.io/github/issues/ooguz/neo146) ![License](https://img.shields.io/github/license/ooguz/neo146) <a href="https://www.buymeacoffee.com/ooguz" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: 20px !important;width: 85px !important;" ></a>

## About

**neo146** provides a minimal (and experimental!) information gateway that serves as an emergency network connection method inspired by dial-up, allowing you to access content via certain protocols. The current implementations are HTTP-SMS gateway and HTTP-Markdown gateway.

The name **"neo146"** comes from the Turkish historic public dial-up service, which operated on dial number 146. This project is created within a few days to support the protesters during the 19 March 2025 uprising in Turkey, following the arrest of the mayor of Istanbul, Ekrem Imamoglu who is the opposition's presidential candidate.

The service is free, but running it costs about 20 EUR per month, and also 5-20 cents per message for the SMS gateway; please use responsibly. For supporting the service and a better experience, please consider donating and subscribing.

SMS responses are base64 encoded for using less SMS credits. Multiple messages are used to send longer responses, the sequence of messages is indicated in the response as `GW<number>|` prefix.

HTTP responses are not encoded by default, but can be requested with `b64=true` parameter.

```
+-------------------+
| +90 850 242 0 146 |
+-------------------+
```

## Available Commands

### SMS Commands
*   `URL (https://...)` - Fetch and convert any webpage to Markdown format
*   `twitter user <username>` - Get the last 5 tweets from a Twitter user
*   `websearch <query>` - Search the web using DuckDuckGo
*   `wiki <2charlangcode> <query>` - Get Wikipedia article summary
*   `weather <location>` - Get weather forecast for a location

### Telegram Bot Commands
*   `/url <url>` - Convert webpage to Markdown format
*   `/twitter <username>` - Get last 5 tweets from a Twitter user
*   `/search <query>` - Search the web using DuckDuckGo
*   `/wiki <lang> <query>` - Get Wikipedia article summary
*   `/weather <location>` - Get weather forecast for a location
*   `/subscribe <email>` - Subscribe to the service
*   `/help` - Show available commands

## HTTP Endpoints

*   `/uri2md?uri=<uri>[&b64=true]` - Convert URI to Markdown
*   `/twitter?user=<user>[&b64=true]` - Get last 5 tweets of a user
*   `/ddg?q=<query>[&b64=true]` - Search the web via DuckDuckGo
*   `/wiki?lang=<2charlangcode>&q=<query>[&b64=true]` - Get Wikipedia article summary
*   `/weather?loc=<location>` - Get weather forecast

## Rate Limits

*   SMS: 5 messages per hour per phone number
*   Telegram: 5 messages per hour per user
*   Subscribe to support the service and get 20 messages/hour

## Subscription

*   Get higher rate limits by subscribing via: [Buy Me a Coffee](https://buymeacoffee.com/ooguz)
*   After subscribing, use `/subscribe <your-email>` in Telegram or text "subscribe <your-email>" to the SMS number
*   All your contribution will be used to maintain the service, rest will be donated to Free Software Association in Turkey (Özgür Yazılım Derneği) [https://oyd.org.tr](https://oyd.org.tr)
*   Running this service costs about 20 EUR per month, and also ~3 cents/message for the SMS gateway. For a better experience and support the service, please consider subscribing.

## Donations

### Monetary donations

*   [Buy Me a Coffee](https://buymeacoffee.com/ooguz)
*   [PayPal](https://paypal.me/ozcanoguz)
    

### Hardware and other donations needed

*   Satellite hardware (VSAT, Starlink, Viasat etc.)
*   Dial-up modems or telephones (DTMF)
*   LoRa and LoRaWAN devices
*   Amateur radio hardware: Radios, antennas, Signalink(-like) devices
*   VoIP numbers, bulk SMS subscriptions or credits
*   Servers, any kind
*   Any service credits useful for running the service

## Support and Contact

*   Email: neo146 \[at\] riseup \[dot\] net (preferred)
*   Twitter: [@ooguz](https://twitter.com/ooguz)
*   Mailing list: [neo146-users](https://lists.riseup.net/www/subscribe/neo146-users)
*   Follow us on Twitter: [@neo146net](https://twitter.com/neo146net)
*   Follow us on Mastodon: [@neo146@chaos.social](https://chaos.social/@neo146)
*   Telegram Channel: [https://t.me/neo146net](https://t.me/neo146net)

## Thanks

*   [wttr.in](https://wttr.in) - weather data
*   [DuckDuckGo Lite](https://lite.duckduckgo.com/lite) - search engine
*   [urltomarkdown](https://github.com/macsplit/urltomarkdown) - markdown conversion
*   [Nitter project](https://github.com/zedeus/nitter) - Twitter API
*   [Özgür Yazılım Derneği](https://oyd.org.tr) - support

This gateway is free software, licensed under GNU AGPL v3 or later. 

## Warning

*   This service is provided as-is, without any warranty. Use at your own risk.
*   The service is not responsible for any content accessed via the gateway.
*   SMS messages are not encrypted — do not use for sensitive content. Your messages may be read by the provider or government.
*   This is a personal, non-commercial project. Subscriptions are for support, not business.
*   Please do not abuse the service by sending spam or malicious content.
*   The service is not affiliated with any organization. It is a personal project.
