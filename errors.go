package textfully

import (
	"fmt"
)

// APIError represents an error returned by the API
type APIError struct {
	StatusCode int    `json:"-"`        // HTTP status code
	Type       string `json:"type"`     // Error type
	Message    string `json:"message"`  // Error message
	Response   string `json:"response"` // Raw response body (if parsing failed)
}

func (e *APIError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("textfully: %s (Type: %s)", e.Message, e.Type)
	}
	return fmt.Sprintf("textfully: %s", e.Message)
}
