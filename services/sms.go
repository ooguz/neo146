package services

import (
	"fmt"
	"neo146/providers"
	"neo146/utils"
	"os"
	"time"
)

// SMSService handles SMS operations
type SMSService struct {
	ProviderManager *providers.Manager
	MarkdownService *MarkdownService
	TwitterService  *TwitterService
	SearchService   *SearchService
	WeatherService  *WeatherService
}

// NewSMSService creates a new SMS service
func NewSMSService(providerManager *providers.Manager) *SMSService {
	return &SMSService{
		ProviderManager: providerManager,
	}
}

// SendSMS sends SMS messages through the SMS provider
func (s *SMSService) SendSMS(messages []providers.Message) error {
	return s.ProviderManager.SendMessage(messages)
}

// PrepareAndSendSMS prepares and sends SMS messages
func (s *SMSService) PrepareAndSendSMS(content string, destinationAddr string, encode bool) error {
	var smsMessages []providers.Message

	if encode {
		// Split and encode the message
		encodedParts := utils.SplitAndEncodeMessage(content, 500)

		// Create response messages
		for i, encoded := range encodedParts {
			smsMessages = append(smsMessages, providers.Message{
				Msg:  encoded,
				Dest: destinationAddr,
				ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), i),
			})
		}
	} else {
		// Send without encoding
		smsMessages = append(smsMessages, providers.Message{
			Msg:  content,
			Dest: destinationAddr,
			ID:   fmt.Sprintf("%d_%d", time.Now().Unix(), 0),
		})
	}

	return s.SendSMS(smsMessages)
}

// CreateSMSRequest creates an SMS request payload
func (s *SMSService) CreateSMSRequest(messages []providers.Message) map[string]interface{} {
	return map[string]interface{}{
		"username":    os.Getenv("SMS_USERNAME"),
		"password":    os.Getenv("SMS_PASSWORD"),
		"source_addr": os.Getenv("SMS_SOURCE_ADDR"),
		"valid_for":   "48:00",
		"datacoding":  "0",
		"messages":    messages,
	}
}
