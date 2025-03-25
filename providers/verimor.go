package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// VerimorProvider implements the Provider interface for Verimor SMS service
type VerimorProvider struct {
	client *http.Client
}

// NewVerimorProvider creates a new VerimorProvider instance
func NewVerimorProvider() *VerimorProvider {
	return &VerimorProvider{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name returns the name of the provider
func (v *VerimorProvider) Name() string {
	return "Verimor"
}

// SMSRequest represents the request structure for Verimor API
type SMSRequest struct {
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	SourceAddr string    `json:"source_addr"`
	ValidFor   string    `json:"valid_for"`
	SendAt     string    `json:"send_at"`
	CustomID   string    `json:"custom_id"`
	Datacoding string    `json:"datacoding"`
	Messages   []Message `json:"messages"`
}

// Send sends one or more SMS messages using Verimor's API
func (v *VerimorProvider) Send(messages []Message) error {
	username := os.Getenv("SMS_USERNAME")
	password := os.Getenv("SMS_PASSWORD")
	sourceAddr := os.Getenv("SMS_SOURCE_ADDR")

	// Validate required credentials
	if username == "" || password == "" || sourceAddr == "" {
		return fmt.Errorf("missing required SMS credentials (SMS_USERNAME, SMS_PASSWORD, or SMS_SOURCE_ADDR)")
	}

	// Validate messages
	if len(messages) == 0 {
		return fmt.Errorf("no messages to send")
	}

	for _, msg := range messages {
		if msg.Dest == "" {
			return fmt.Errorf("destination address cannot be empty")
		}
	}

	datacoding := "0"
	if messages[0].Msg[0] != 'G' {
		datacoding = "2"
	}

	smsRequest := SMSRequest{
		Username:   username,
		Password:   password,
		SourceAddr: sourceAddr,
		ValidFor:   "48:00",
		Datacoding: datacoding,
		Messages:   messages,
	}

	jsonData, err := json.Marshal(smsRequest)
	if err != nil {
		return fmt.Errorf("error marshaling SMS request: %v", err)
	}

	// Log sent payload
	fmt.Printf("Sending SMS payload:\n%s\n", string(jsonData))

	resp, err := v.client.Post(
		"https://sms.verimor.com.tr/v2/send.json",
		"application/json",
		strings.NewReader(string(jsonData)),
	)
	if err != nil {
		return fmt.Errorf("error sending SMS: %v", err)
	}
	defer resp.Body.Close()

	// Log response
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("SMS API Response (Status: %d):\n%s\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SMS API error: %s", string(body))
	}

	return nil
}
