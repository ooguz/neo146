//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func main() {
	// Test with a few different queries
	queries := []string{"golang programming", "climate change", "artificial intelligence"}

	for _, query := range queries {
		fmt.Printf("\n\nTesting search for: %s\n", query)
		results, err := fetchDuckDuckGoResults(query)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println("Results found:")
		fmt.Println("======================================")
		fmt.Println(results)
		fmt.Println("======================================")
	}
}

// fetchDuckDuckGoResults performs a search on DuckDuckGo lite and returns the results
func fetchDuckDuckGoResults(query string) (string, error) {
	// Create form data for the POST request
	formData := url.Values{}
	formData.Set("q", query)
	formData.Set("kl", "us-en") // US English results

	// DuckDuckGo Lite expects a POST request with form data
	resp, err := httpClient.PostForm("https://lite.duckduckgo.com/lite/", formData)
	if err != nil {
		return "", fmt.Errorf("error fetching DuckDuckGo results: %v", err)
	}
	defer resp.Body.Close()

	// Save the raw HTML for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	htmlContent := string(body)

	// Debug - log response status and URL
	fmt.Printf("DuckDuckGo search returned status: %d\n", resp.StatusCode)
	fmt.Printf("URL: %s\n", resp.Request.URL)

	// Extract search results using multiple approaches
	var results []string

	// Approach 1: Direct HTML parsing with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err == nil {
		// First, check if we got any search results
		fmt.Printf("Found %d tables in the response\n", doc.Find("table").Length())
		fmt.Printf("Found %d links in the response\n", doc.Find("a").Length())

		// Find all links and extract those that point to external sites
		doc.Find("a").Each(func(i int, link *goquery.Selection) {
			href, exists := link.Attr("href")
			// Skip DuckDuckGo internal links
			if exists && !strings.Contains(href, "duckduckgo.com") &&
				strings.HasPrefix(href, "http") && link.Text() != "" {
				title := strings.TrimSpace(link.Text())
				fmt.Printf("Found link: '%s' -> %s\n", title, href)

				// Skip navigation links like "Next" or very short titles
				if len(title) > 5 &&
					!strings.EqualFold(title, "Next") &&
					!strings.EqualFold(title, "Previous") {
					results = append(results, fmt.Sprintf("# %s\n%s", title, href))
				}
			}
		})
	}

	// If the first approach found no results, try a different one
	if len(results) == 0 {
		fmt.Println("No results from first approach, trying tables...")

		// Look specifically for result patterns - in DuckDuckGo Lite results are usually in tables
		if doc != nil {
			doc.Find("table").Each(func(i int, table *goquery.Selection) {
				fmt.Printf("Checking table %d, has %d rows\n", i, table.Find("tr").Length())

				// Look for tables that contain search results - usually have multiple rows
				if table.Find("tr").Length() > 1 {
					table.Find("tr").Each(func(j int, row *goquery.Selection) {
						// Only process rows with links
						if row.Find("a").Length() > 0 {
							link := row.Find("a").First()
							href, exists := link.Attr("href")
							if exists && !strings.Contains(href, "duckduckgo.com") &&
								strings.HasPrefix(href, "http") {
								title := strings.TrimSpace(link.Text())
								if title != "" {
									fmt.Printf("Found result in table: '%s' -> %s\n", title, href)
									results = append(results, fmt.Sprintf("# %s\n%s", title, href))
								}
							}
						}
					})
				}
			})
		}
	}

	// If still no results, try direct regex extraction
	if len(results) == 0 {
		fmt.Println("Using regex to extract links from HTML...")
		links := extractLinksFromHTML(htmlContent)
		fmt.Printf("Found %d links with regex\n", len(links))

		for _, link := range links {
			if !strings.Contains(link, "duckduckgo.com") {
				results = append(results, fmt.Sprintf("# Search Result\n%s", link))
			}
		}
	}

	// Final check for debugging
	if len(results) == 0 {
		// Log the HTML structure to help diagnose the issue
		fmt.Println("No results found after all attempts")
		// Save a portion of the HTML for debugging
		if len(htmlContent) > 300 {
			fmt.Printf("HTML (first 300 chars): %s\n", htmlContent[:300])
		} else {
			fmt.Printf("HTML: %s\n", htmlContent)
		}
		return "No search results found. Please try a different query.", nil
	}

	// Limit to 5 results maximum for SMS
	fmt.Printf("Found %d total results\n", len(results))
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
