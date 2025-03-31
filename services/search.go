package services

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// SearchService handles searching using DuckDuckGo
type SearchService struct {
	httpClient *http.Client
}

// NewSearchService creates a new instance of SearchService
func NewSearchService(httpClient *http.Client) *SearchService {
	return &SearchService{
		httpClient: httpClient,
	}
}

// FetchDuckDuckGoResults fetches search results from DuckDuckGo
func (s *SearchService) FetchDuckDuckGoResults(query string) (string, error) {
	url := fmt.Sprintf("https://lite.duckduckgo.com/lite/?q=%s", query)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse the HTML to extract search results
	results := s.parseSearchResults(string(body))
	if results == "" {
		return "No results found", nil
	}

	return results, nil
}

// parseSearchResults parses HTML from DuckDuckGo to extract search results
func (s *SearchService) parseSearchResults(html string) string {
	// This is a very basic parser that extracts links and snippets
	// A better approach would be to use an HTML parser like goquery

	// Extract links
	linkRegex := regexp.MustCompile(`<a rel="nofollow" href="([^"]+)"[^>]*>([^<]+)</a>`)
	matches := linkRegex.FindAllStringSubmatch(html, 10) // Limit to 10 results

	var results []string
	for _, match := range matches {
		if len(match) >= 3 {
			url := match[1]
			title := match[2]

			// Skip ads and irrelevant links
			if strings.Contains(url, "duckduckgo.com") {
				continue
			}

			results = append(results, fmt.Sprintf("- %s\n  %s", title, url))
		}
	}

	return strings.Join(results, "\n\n")
}

// FetchWikipediaSummary fetches a summary from Wikipedia
func (s *SearchService) FetchWikipediaSummary(query string, langCode string) (string, error) {
	url := fmt.Sprintf("https://%s.wikipedia.org/w/api.php?format=json&action=query&prop=extracts&exintro&explaintext&redirects=1&titles=%s",
		langCode, query)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse the JSON response to extract the summary
	// This is a simplified implementation
	summary := string(body)

	// In a real implementation, you would parse the JSON and extract the actual summary

	return summary, nil
}
