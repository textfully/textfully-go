package textfully

import "time"

// MessageRequest represents a text message to be sent
type MessageRequest struct {
	PhoneNumber string `json:"phone_number"`
	Text        string `json:"text"`
}

// MessageResponse represents the API response for a sent message
type MessageResponse struct {
	ID        string    `json:"id"`         // Unique message ID
	Status    string    `json:"status"`     // Message status (queued, sent, delivered, failed)
	CreatedAt time.Time `json:"created_at"` // Timestamp when message was created
}
