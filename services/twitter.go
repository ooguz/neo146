package services

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"smsgw/models"
	"strings"
)

// TwitterService handles fetching tweets
type TwitterService struct {
	httpClient *http.Client
}

// NewTwitterService creates a new instance of TwitterService
func NewTwitterService(httpClient *http.Client) *TwitterService {
	return &TwitterService{
		httpClient: httpClient,
	}
}

// FetchTweets fetches tweets for a user
func (s *TwitterService) FetchTweets(username string, count int) (string, error) {
	nitterURL := fmt.Sprintf("https://nitter.app.ooguz.dev/%s/rss", username)
	resp, err := s.httpClient.Get(nitterURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tweets, err := s.parseTweetsFromRSS(string(body), count)
	if err != nil {
		return "", err
	}

	return tweets, nil
}

// parseTweetsFromRSS parses tweets from an RSS feed
func (s *TwitterService) parseTweetsFromRSS(rss string, count int) (string, error) {
	var feed models.RSS
	if err := xml.Unmarshal([]byte(rss), &feed); err != nil {
		return "", fmt.Errorf("error parsing RSS: %v", err)
	}

	if len(feed.Channel.Items) == 0 {
		return "", fmt.Errorf("no tweets found")
	}

	var tweets []string
	// Process items until we have enough non-retweet tweets
	for i := 0; i < len(feed.Channel.Items) && len(tweets) < count; i++ {
		item := feed.Channel.Items[i]

		// Clean up HTML entities
		title := item.Title
		title = strings.ReplaceAll(title, "&quot;", "\"")
		title = strings.ReplaceAll(title, "&apos;", "'")
		title = strings.ReplaceAll(title, "&lt;", "<")
		title = strings.ReplaceAll(title, "&gt;", ">")
		title = strings.ReplaceAll(title, "&amp;", "&")

		// Skip RT by @username: tweets
		if strings.HasPrefix(title, "RT by @") {
			continue
		}

		// Skip empty titles
		if strings.TrimSpace(title) == "" {
			continue
		}

		// Add the tweet
		tweets = append(tweets, fmt.Sprintf("- %s", title))
	}

	if len(tweets) == 0 {
		return "", fmt.Errorf("no tweets found")
	}

	return strings.Join(tweets, "\n"), nil
}
