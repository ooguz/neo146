package services

import (
	"fmt"
	"io"
	"net/http"
)

// MarkdownService handles fetching and converting content to Markdown
type MarkdownService struct {
	httpClient *http.Client
}

// NewMarkdownService creates a new instance of MarkdownService
func NewMarkdownService(httpClient *http.Client) *MarkdownService {
	return &MarkdownService{
		httpClient: httpClient,
	}
}

// FetchMarkdown fetches a URL and converts it to Markdown
func (s *MarkdownService) FetchMarkdown(url string) (string, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("https://urltomarkdown.herokuapp.com/?clean=true&url=%s", url))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	markdown, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(markdown), nil
}
