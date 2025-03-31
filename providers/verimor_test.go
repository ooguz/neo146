package providers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestVerimorProvider_Name(t *testing.T) {
	provider := NewVerimorProvider()

	if provider.Name() != "Verimor" {
		t.Errorf("Expected provider name to be Verimor, got %s", provider.Name())
	}
}

func TestVerimorProvider_Send(t *testing.T) {
	// Save original env vars to restore later
	originalUsername := os.Getenv("SMS_USERNAME")
	originalPassword := os.Getenv("SMS_PASSWORD")
	originalSourceAddr := os.Getenv("SMS_SOURCE_ADDR")

	defer func() {
		os.Setenv("SMS_USERNAME", originalUsername)
		os.Setenv("SMS_PASSWORD", originalPassword)
		os.Setenv("SMS_SOURCE_ADDR", originalSourceAddr)
	}()

	// Set test env vars
	os.Setenv("SMS_USERNAME", "testuser")
	os.Setenv("SMS_PASSWORD", "testpass")
	os.Setenv("SMS_SOURCE_ADDR", "TESTSRC")

	// Create test messages
	messages := []Message{
		{Msg: "Test message 1", Dest: "+1234567890", ID: "1"},
		{Msg: "Test message 2", Dest: "+0987654321", ID: "2"},
	}

	// Test server to mock Verimor API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/send.json" {
			t.Errorf("Expected request to /v2/send.json, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"message_count": 2, "successful": true}`)
	}))
	defer server.Close()

	// Create provider with custom client that uses our test server
	provider := &VerimorProvider{
		client: &http.Client{
			Transport: &customTransport{
				originalTransport: server.Client().Transport,
				baseURL:           server.URL,
			},
		},
	}

	// Test successful send
	err := provider.Send(messages)
	if err != nil {
		t.Errorf("Expected successful message send, got error: %v", err)
	}

	// Test missing credentials
	os.Unsetenv("SMS_USERNAME")
	err = provider.Send(messages)
	if err == nil {
		t.Error("Expected error when missing username, got nil")
	}

	os.Setenv("SMS_USERNAME", "testuser")
	os.Unsetenv("SMS_PASSWORD")
	err = provider.Send(messages)
	if err == nil {
		t.Error("Expected error when missing password, got nil")
	}

	os.Setenv("SMS_PASSWORD", "testpass")
	os.Unsetenv("SMS_SOURCE_ADDR")
	err = provider.Send(messages)
	if err == nil {
		t.Error("Expected error when missing source address, got nil")
	}

	// Test empty message list
	os.Setenv("SMS_SOURCE_ADDR", "TESTSRC")
	err = provider.Send([]Message{})
	if err == nil {
		t.Error("Expected error when sending empty message list, got nil")
	}

	// Test message with empty destination
	err = provider.Send([]Message{{Msg: "Test", Dest: "", ID: "1"}})
	if err == nil {
		t.Error("Expected error when sending message with empty destination, got nil")
	}

	// Test server error response
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `{"error": "Invalid request"}`)
	}))
	defer errorServer.Close()

	// Create a new provider with error server
	errorProvider := &VerimorProvider{
		client: &http.Client{
			Transport: &customTransport{
				originalTransport: errorServer.Client().Transport,
				baseURL:           errorServer.URL,
			},
		},
	}

	err = errorProvider.Send(messages)
	if err == nil {
		t.Error("Expected error when server returns error status, got nil")
	}
}

// customTransport is a http.RoundTripper that modifies the request URL
type customTransport struct {
	originalTransport http.RoundTripper
	baseURL           string
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Save the original request URL
	originalURL := req.URL

	// Create a new URL based on our test server
	req.URL.Scheme = "http"
	req.URL.Host = t.baseURL[7:] // Remove "http://" prefix
	req.Host = req.URL.Host

	// Preserve the path but override the scheme and host
	req.URL.Path = "/v2/send.json"

	// Make the request
	resp, err := t.originalTransport.RoundTrip(req)

	// Restore the original URL (cleanup)
	req.URL = originalURL

	return resp, err
}
