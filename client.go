package textfully

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"
)

const (
	defaultBaseURL  = "https://api.textfully.dev/v1"
	defaultTimeout  = 30 * time.Second
	e164PhoneRegexp = `^\+[1-9]\d{1,14}$`
)

var (
	// phoneRegexp is used to validate E.164 phone numbers
	phoneRegexp = regexp.MustCompile(e164PhoneRegexp)
)

// Client represents a Textfully API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// New creates a new Textfully client with the given API key
func New(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Send sends a text message
func (c *Client) Send(phoneNumber, text string) (*MessageResponse, error) {
	// Validate API key
	if c.apiKey == "" {
		return nil, &APIError{
			StatusCode: http.StatusUnauthorized,
			Message:    "No API key provided. Set your API key using textfully.New('tx_apikey')",
		}
	}

	// Validate phone number format
	if !phoneRegexp.MatchString(phoneNumber) {
		return nil, fmt.Errorf("invalid phone number format. Must be in E.164 format (e.g., +16175555555)")
	}

	// Create message request
	message := &MessageRequest{
		PhoneNumber: phoneNumber,
		Text:        text,
	}

	// Marshal request to JSON
	payload, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create request
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/messages", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "textfully-go/"+Version)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
			return nil, &APIError{
				Type:    "timeout_error",
				Message: "Request timed out. Please try again.",
			}
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var result MessageResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    "Invalid JSON response from API",
				Response:   err.Error(),
			}
		}
		return &result, nil
	}

	// Handle error response
	var apiError struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiError); err != nil {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to decode error response",
			Response:   err.Error(),
		}
	}

	err = &APIError{
		StatusCode: resp.StatusCode,
		Type:       apiError.Error.Type,
		Message:    apiError.Error.Message,
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("authentication failed: %w", err)
	case http.StatusBadRequest:
		return nil, fmt.Errorf("bad request: %w", err)
	default:
		return nil, fmt.Errorf("API request failed: %w", err)
	}
}
