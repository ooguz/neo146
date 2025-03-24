package providers

// Message represents an SMS message to be sent
type Message struct {
	Msg  string
	Dest string
	ID   string
}

// Provider defines the interface that all SMS providers must implement
type Provider interface {
	// Send sends one or more SMS messages
	Send(messages []Message) error
	// Name returns the name of the provider
	Name() string
}
