package providers

import (
	"fmt"
	"os"
	"sync"
)

// Manager handles SMS providers
type Manager struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewManager creates a new provider manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]Provider),
	}
}

// RegisterProvider registers a new SMS provider
func (m *Manager) RegisterProvider(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.Name()] = provider
}

// GetProvider returns a provider by name
func (m *Manager) GetProvider(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// SendMessage sends a message using the configured provider
func (m *Manager) SendMessage(messages []Message) error {
	providerName := os.Getenv("SMS_PROVIDER")
	if providerName == "" {
		providerName = "Verimor" // Default provider
	}

	provider, err := m.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("error getting provider: %v", err)
	}

	return provider.Send(messages)
}
