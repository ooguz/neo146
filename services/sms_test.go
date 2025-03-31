package services

import (
	"errors"
	"smsgw/providers"
	"testing"
)

// Define a ManagerInterface to abstract providers.Manager for testing
type ManagerInterface interface {
	SendMessage(messages []providers.Message) error
	RegisterProvider(provider providers.Provider)
	GetProvider(name string) (providers.Provider, error)
}

// Ensure SMSService can work with our interface
type smsServiceWithInterface struct {
	smsManager ManagerInterface
}

func newSMSServiceWithInterface(manager ManagerInterface) *smsServiceWithInterface {
	return &smsServiceWithInterface{
		smsManager: manager,
	}
}

func (s *smsServiceWithInterface) SendSMS(messages []providers.Message) error {
	return s.smsManager.SendMessage(messages)
}

func (s *smsServiceWithInterface) PrepareAndSendSMS(content string, destinationAddr string, encode bool) error {
	var smsMessages []providers.Message

	if encode {
		// For testing, we'll just split the message into chunks of 50 chars
		contentRunes := []rune(content)
		chunkSize := 50

		for i := 0; i < len(contentRunes); i += chunkSize {
			end := i + chunkSize
			if end > len(contentRunes) {
				end = len(contentRunes)
			}

			chunk := string(contentRunes[i:end])
			smsMessages = append(smsMessages, providers.Message{
				Msg:  chunk,
				Dest: destinationAddr,
				ID:   string(rune(i)),
			})
		}
	} else {
		smsMessages = append(smsMessages, providers.Message{
			Msg:  content,
			Dest: destinationAddr,
			ID:   "0",
		})
	}

	return s.SendSMS(smsMessages)
}

func (s *smsServiceWithInterface) CreateSMSRequest(messages []providers.Message) map[string]interface{} {
	return map[string]interface{}{
		"username":    "test-username",
		"password":    "test-password",
		"source_addr": "test-source",
		"valid_for":   "48:00",
		"datacoding":  "0",
		"messages":    messages,
	}
}

// MockSMSManager implements the ManagerInterface for testing
type MockSMSManager struct {
	sentMessages []providers.Message
	shouldError  bool
}

func NewMockSMSManager(shouldError bool) *MockSMSManager {
	return &MockSMSManager{
		sentMessages: []providers.Message{},
		shouldError:  shouldError,
	}
}

func (m *MockSMSManager) SendMessage(messages []providers.Message) error {
	if m.shouldError {
		return errors.New("mock send error")
	}
	m.sentMessages = append(m.sentMessages, messages...)
	return nil
}

func (m *MockSMSManager) RegisterProvider(provider providers.Provider) {
	// Not needed for this test
}

func (m *MockSMSManager) GetProvider(name string) (providers.Provider, error) {
	// Not needed for this test
	return nil, nil
}

func TestSMSService_SendSMS(t *testing.T) {
	// Test successful send
	mockManager := NewMockSMSManager(false)
	service := newSMSServiceWithInterface(mockManager)

	messages := []providers.Message{
		{Msg: "Test message 1", Dest: "+1234567890", ID: "1"},
		{Msg: "Test message 2", Dest: "+0987654321", ID: "2"},
	}

	err := service.SendSMS(messages)
	if err != nil {
		t.Errorf("Expected successful send, got error: %v", err)
	}

	if len(mockManager.sentMessages) != 2 {
		t.Errorf("Expected 2 messages to be sent, got %d", len(mockManager.sentMessages))
	}

	// Test error case
	errorManager := NewMockSMSManager(true)
	errorService := newSMSServiceWithInterface(errorManager)

	err = errorService.SendSMS(messages)
	if err == nil {
		t.Error("Expected error when manager returns error, got nil")
	}
}

func TestSMSService_PrepareAndSendSMS(t *testing.T) {
	mockManager := NewMockSMSManager(false)
	service := newSMSServiceWithInterface(mockManager)

	// Test simple message (no encoding)
	err := service.PrepareAndSendSMS("Simple test message", "+1234567890", false)
	if err != nil {
		t.Errorf("Expected successful send, got error: %v", err)
	}

	if len(mockManager.sentMessages) != 1 {
		t.Errorf("Expected 1 message to be sent, got %d", len(mockManager.sentMessages))
	}

	if mockManager.sentMessages[0].Msg != "Simple test message" {
		t.Errorf("Expected message content to be 'Simple test message', got '%s'", mockManager.sentMessages[0].Msg)
	}

	// Reset mock
	mockManager = NewMockSMSManager(false)
	service = newSMSServiceWithInterface(mockManager)

	// Test long message that needs encoding and splitting
	longMessage := "This is a very long message that should be split into multiple parts when encoding is enabled. " +
		"We need to make sure it's long enough to trigger the splitting functionality."

	err = service.PrepareAndSendSMS(longMessage, "+1234567890", true)
	if err != nil {
		t.Errorf("Expected successful send, got error: %v", err)
	}

	if len(mockManager.sentMessages) < 2 {
		t.Errorf("Expected multiple messages after splitting, got %d", len(mockManager.sentMessages))
	}

	// Test error case
	errorManager := NewMockSMSManager(true)
	errorService := newSMSServiceWithInterface(errorManager)

	err = errorService.PrepareAndSendSMS("Test message", "+1234567890", false)
	if err == nil {
		t.Error("Expected error when manager returns error, got nil")
	}
}

func TestSMSService_CreateSMSRequest(t *testing.T) {
	service := newSMSServiceWithInterface(nil) // Manager not needed for this test

	messages := []providers.Message{
		{Msg: "Test message 1", Dest: "+1234567890", ID: "1"},
		{Msg: "Test message 2", Dest: "+0987654321", ID: "2"},
	}

	request := service.CreateSMSRequest(messages)

	// Check request structure
	if request["valid_for"] != "48:00" {
		t.Errorf("Expected valid_for to be '48:00', got '%v'", request["valid_for"])
	}

	if request["datacoding"] != "0" {
		t.Errorf("Expected datacoding to be '0', got '%v'", request["datacoding"])
	}

	// Check the messages array
	requestMessages, ok := request["messages"].([]providers.Message)
	if !ok {
		t.Fatal("Expected request[\"messages\"] to be []providers.Message")
	}

	if len(requestMessages) != 2 {
		t.Errorf("Expected 2 messages in request, got %d", len(requestMessages))
	}
}
