package models

// SMSPayload represents an incoming SMS message
type SMSPayload struct {
	MessageID       int    `json:"message_id"`
	Type            string `json:"type"`
	CreatedAt       string `json:"created_at"`
	Network         string `json:"network"`
	SourceAddr      string `json:"source_addr"`
	DestinationAddr string `json:"destination_addr"`
	Keyword         string `json:"keyword"`
	Content         string `json:"content"`
	ReceivedAt      string `json:"received_at"`
}

// SMSRequest represents a request to send SMS messages
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

// Message represents a single SMS message
type Message struct {
	Msg  string `json:"msg"`
	Dest string `json:"dest"`
	ID   string `json:"id"`
}
