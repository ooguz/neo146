package services

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// WeatherService handles fetching weather data
type WeatherService struct {
	httpClient *http.Client
}

// NewWeatherService creates a new instance of WeatherService
func NewWeatherService(httpClient *http.Client) *WeatherService {
	return &WeatherService{
		httpClient: httpClient,
	}
}

// FetchWeatherForecast fetches weather forecast for a location
func (s *WeatherService) FetchWeatherForecast(location string) (string, error) {
	// Format location for wttr.in
	encodedLocation := url.QueryEscape(location)
	format := url.QueryEscape("%l:\n%c%t\n%w %h - %m\nsr %S\nss %s\n")

	// Call wttr.in with format=3 for a compact forecast
	wttrURL := fmt.Sprintf("https://wttr.in/%s?format=%s",
		url.PathEscape(encodedLocation), format)

	resp, err := s.httpClient.Get(wttrURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
