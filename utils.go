package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// fetchDuckDuckGoResults performs a search on DuckDuckGo lite and returns the results
func fetchDuckDuckGoResults(query string) (string, error) {
	// Create form data for the POST request
	formData := url.Values{}
	formData.Set("q", query)
	formData.Set("kl", "tr-tr") // TR results

	// DuckDuckGo Lite expects a POST request with form data
	resp, err := httpClient.PostForm("https://lite.duckduckgo.com/lite/", formData)
	if err != nil {
		return "", fmt.Errorf("error fetching DuckDuckGo results: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	htmlContent := string(body)

	// Extract search results
	var results []string

	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err == nil {
		// Find all links and extract those that point to external sites
		doc.Find("a").Each(func(i int, link *goquery.Selection) {
			href, exists := link.Attr("href")
			// Skip DuckDuckGo internal links
			if exists && !strings.Contains(href, "duckduckgo.com") &&
				strings.HasPrefix(href, "http") && link.Text() != "" {
				title := strings.TrimSpace(link.Text())

				// Skip navigation links like "Next" or very short titles
				if len(title) > 5 &&
					!strings.EqualFold(title, "Next") &&
					!strings.EqualFold(title, "Previous") {
					results = append(results, fmt.Sprintf("# %s\n%s", title, href))
				}
			}
		})
	}

	// Fallback to regex extraction if no results found
	if len(results) == 0 {
		links := extractLinksFromHTML(htmlContent)
		for _, link := range links {
			if !strings.Contains(link, "duckduckgo.com") {
				results = append(results, fmt.Sprintf("# Search Result\n%s", link))
			}
		}
	}

	// Return message if no results found
	if len(results) == 0 {
		return "No search results found. Please try a different query.", nil
	}

	// Limit to 5 results maximum for SMS
	maxResults := 5
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return strings.Join(results, "\n\n"), nil
}

// Helper function to extract links from HTML using regex
func extractLinksFromHTML(html string) []string {
	var links []string

	// First pattern - standard href links
	re1 := regexp.MustCompile(`href=["'](https?://[^"']+)["']`)
	matches1 := re1.FindAllStringSubmatch(html, -1)

	for _, match := range matches1 {
		if len(match) >= 2 && !strings.Contains(match[1], "duckduckgo.com") {
			links = append(links, match[1])
		}
	}

	// Second pattern - looking for URLs in text
	re2 := regexp.MustCompile(`(https?://[^\s"'<>]+)`)
	matches2 := re2.FindAllStringSubmatch(html, -1)

	for _, match := range matches2 {
		if len(match) >= 2 && !strings.Contains(match[1], "duckduckgo.com") {
			// Check if this URL is already in our list
			isDuplicate := false
			for _, existingLink := range links {
				if existingLink == match[1] {
					isDuplicate = true
					break
				}
			}

			if !isDuplicate {
				links = append(links, match[1])
			}
		}
	}

	return links
}

// fetchWikipediaSummary fetches a summary from Wikipedia in the specified language
func fetchWikipediaSummary(query string, langCode string) (string, error) {
	// Validate language code
	if len(langCode) != 2 {
		return "", fmt.Errorf("invalid language code: %s (should be 2 characters)", langCode)
	}

	// Capitalize the first letter of the query to match Wikipedia's convention
	capitalizedQuery := query
	if len(query) > 0 {
		capitalizedQuery = strings.ToUpper(query[:1]) + query[1:]
	}

	// Create URL for Wikipedia API
	apiURL := fmt.Sprintf("https://%s.wikipedia.org/api/rest_v1/page/summary/%s",
		langCode, url.PathEscape(capitalizedQuery))

	resp, err := httpClient.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("error fetching Wikipedia summary: %v", err)
	}
	defer resp.Body.Close()

	// If the capitalized version fails, try with the original query
	if resp.StatusCode == http.StatusNotFound && capitalizedQuery != query {
		resp.Body.Close() // Close the first response

		// Try with original query
		apiURL = fmt.Sprintf("https://%s.wikipedia.org/api/rest_v1/page/summary/%s",
			langCode, url.PathEscape(query))
		resp, err = httpClient.Get(apiURL)
		if err != nil {
			return "", fmt.Errorf("error fetching Wikipedia summary: %v", err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("Wikipedia article not found for '%s' in language '%s'", query, langCode)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Wikipedia API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading Wikipedia response: %v", err)
	}

	// Parse JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing Wikipedia JSON: %v", err)
	}

	// Extract title and extract
	title, _ := result["title"].(string)
	extract, ok := result["extract"].(string)

	if !ok || extract == "" {
		// If no extract, try the description field as fallback
		description, hasDesc := result["description"].(string)
		if hasDesc && description != "" {
			extract = description
		} else {
			return "", fmt.Errorf("no content found for '%s'", query)
		}
	}

	if title == "" {
		title = query // Use the query as title if none is returned
	}

	return fmt.Sprintf("# %s\n\n%s", title, extract), nil
}

// fetchWeatherForecast fetches weather information from wttr.in for the specified location
func fetchWeatherForecast(location string) (string, error) {
	// Clean and encode the location
	cleanLocation := strings.TrimSpace(location)
	if cleanLocation == "" {
		return "", fmt.Errorf("location cannot be empty")
	}

	// Replace spaces with + and encode the location
	cleanLocation = strings.ReplaceAll(cleanLocation, " ", "+")

	// Format string for wttr.in - shows location, condition, temperature,
	// wind, humidity, moon phase, sunrise and sunset
	format := url.QueryEscape("%l:\n%c%t\n%w+%h+-+%m\nsr+%S\nss+%s\n")

	// Create the URL for the wttr.in API
	weatherURL := fmt.Sprintf("https://wttr.in/%s?format=%s",
		url.PathEscape(cleanLocation), format)

	// Make the request
	resp, err := httpClient.Get(weatherURL)
	if err != nil {
		return "", fmt.Errorf("error fetching weather: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("weather service returned status: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading weather response: %v", err)
	}

	// Return the weather data as is
	return string(body), nil
}
