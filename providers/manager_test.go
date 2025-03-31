package providers

import (
	"errors"
	"os"
	"testing"
)

// MockProvider implements the Provider interface for testing
type MockProvider struct {
	name         string
	sendResponse error
	sentMessages []Message
}

func NewMockProvider(name string, sendResponse error) *MockProvider {
	return &MockProvider{
		name:         name,
		sendResponse: sendResponse,
		sentMessages: []Message{},
	}
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Send(messages []Message) error {
	m.sentMessages = append(m.sentMessages, messages...)
	return m.sendResponse
}

func TestManager_RegisterProvider(t *testing.T) {
	manager := NewManager()
	mockProvider := NewMockProvider("MockSMS", nil)

	manager.RegisterProvider(mockProvider)

	provider, err := manager.GetProvider("MockSMS")
	if err != nil {
		t.Errorf("Expected to find registered provider, got error: %v", err)
	}

	if provider.Name() != "MockSMS" {
		t.Errorf("Expected provider name to be MockSMS, got %s", provider.Name())
	}
}

func TestManager_GetProvider_NotFound(t *testing.T) {
	manager := NewManager()

	_, err := manager.GetProvider("NonExistent")
	if err == nil {
		t.Error("Expected error when getting non-existent provider, got nil")
	}
}

func TestManager_SendMessage(t *testing.T) {
	// Setup
	originalProvider := os.Getenv("SMS_PROVIDER")
	defer os.Setenv("SMS_PROVIDER", originalProvider)

	os.Setenv("SMS_PROVIDER", "MockSMS")

	manager := NewManager()
	mockProvider := NewMockProvider("MockSMS", nil)
	manager.RegisterProvider(mockProvider)

	messages := []Message{
		{Msg: "Test message 1", Dest: "+1234567890", ID: "1"},
		{Msg: "Test message 2", Dest: "+0987654321", ID: "2"},
	}

	// Test successful send
	err := manager.SendMessage(messages)
	if err != nil {
		t.Errorf("Expected successful message send, got error: %v", err)
	}

	if len(mockProvider.sentMessages) != 2 {
		t.Errorf("Expected 2 messages to be sent, got %d", len(mockProvider.sentMessages))
	}

	// Test provider not found
	os.Setenv("SMS_PROVIDER", "NonExistent")
	err = manager.SendMessage(messages)
	if err == nil {
		t.Error("Expected error when sending with non-existent provider, got nil")
	}

	// Test empty provider name (should use default "Verimor")
	os.Setenv("SMS_PROVIDER", "")
	verimor := NewMockProvider("Verimor", nil)
	manager.RegisterProvider(verimor)

	err = manager.SendMessage(messages)
	if err != nil {
		t.Errorf("Expected successful message send with default provider, got error: %v", err)
	}

	// Test provider returns error
	errorProvider := NewMockProvider("ErrorProvider", errors.New("send error"))
	manager.RegisterProvider(errorProvider)
	os.Setenv("SMS_PROVIDER", "ErrorProvider")

	err = manager.SendMessage(messages)
	if err == nil {
		t.Error("Expected error when provider returns error, got nil")
	}
}
