package textfully

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSend(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		phone     string
		text      string
		wantErr   bool
		setupMock func(w http.ResponseWriter)
		checkErr  func(t *testing.T, err error)
	}{
		{
			name:    "successful send",
			apiKey:  "test_key",
			phone:   "+16175555555",
			text:    "Test message",
			wantErr: false,
			setupMock: func(w http.ResponseWriter) {
				resp := MessageResponse{
					ID:        "msg_123",
					Status:    "sent",
					CreatedAt: time.Now(),
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
		},
		{
			name:      "invalid phone number",
			apiKey:    "test_key",
			phone:     "1234567890",
			text:      "Test message",
			wantErr:   true,
			setupMock: func(w http.ResponseWriter) {},
			checkErr: func(t *testing.T, err error) {
				if !strings.Contains(err.Error(), "invalid phone number format") {
					t.Errorf("Expected invalid phone number error, got: %v", err)
				}
			},
		},
		{
			name:      "missing api key",
			apiKey:    "",
			phone:     "+16175555555",
			text:      "Test message",
			wantErr:   true,
			setupMock: func(w http.ResponseWriter) {},
			checkErr: func(t *testing.T, err error) {
				if !strings.Contains(err.Error(), "No API key provided") {
					t.Errorf("Expected no API key error, got: %v", err)
				}
			},
		},
		{
			name:    "authentication error",
			apiKey:  "invalid_key",
			phone:   "+16175555555",
			text:    "Test message",
			wantErr: true,
			setupMock: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]string{
						"type":    "authentication_error",
						"message": "Invalid API key",
					},
				})
			},
			checkErr: func(t *testing.T, err error) {
				if !strings.Contains(err.Error(), "authentication failed") {
					t.Errorf("Expected authentication error, got: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify HTTP method
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				// Verify content type
				if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}

				// Verify user agent
				if userAgent := r.Header.Get("User-Agent"); !strings.HasPrefix(userAgent, "textfully-go/") {
					t.Errorf("Expected User-Agent to start with textfully-go/, got %s", userAgent)
				}

				// Verify auth header if api key provided
				if tt.apiKey != "" {
					if auth := r.Header.Get("Authorization"); auth != "Bearer "+tt.apiKey {
						t.Errorf("Expected Authorization header 'Bearer %s', got '%s'", tt.apiKey, auth)
					}
				}

				// Verify request body
				if r.Body != nil {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatalf("Failed to read request body: %v", err)
					}
					var reqBody MessageRequest
					if err := json.Unmarshal(body, &reqBody); err != nil {
						t.Fatalf("Failed to parse request body: %v", err)
					}
					if reqBody.PhoneNumber != tt.phone {
						t.Errorf("Expected phone number %s, got %s", tt.phone, reqBody.PhoneNumber)
					}
					if reqBody.Text != tt.text {
						t.Errorf("Expected text %s, got %s", tt.text, reqBody.Text)
					}
				}

				tt.setupMock(w)
			}))
			defer server.Close()

			client := New(tt.apiKey)
			client.baseURL = server.URL

			resp, err := client.Send(tt.phone, tt.text)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp == nil {
					t.Error("Expected response but got nil")
				} else if resp.ID != "msg_123" {
					t.Errorf("Expected message ID 'msg_123', got '%s'", resp.ID)
				}
			}
		})
	}
}

func TestSendRequestCreation(t *testing.T) {
	tests := []struct {
		name    string
		message interface{} // Use interface{} to test various types
		wantErr string
	}{
		{
			name: "invalid message - channel",
			message: struct {
				PhoneNumber string   `json:"phone_number"`
				Text        chan int `json:"text"`
			}{
				PhoneNumber: "+16175555555",
				Text:        make(chan int),
			},
			wantErr: "json: unsupported type: chan int",
		},
		{
			name: "invalid message - function",
			message: struct {
				PhoneNumber string `json:"phone_number"`
				Text        func() `json:"text"`
			}{
				PhoneNumber: "+16175555555",
				Text:        func() {},
			},
			wantErr: "json: unsupported type: func()",
		},
		{
			name: "invalid message - complex type",
			message: struct {
				PhoneNumber string           `json:"phone_number"`
				Text        map[chan int]int `json:"text"`
			}{
				PhoneNumber: "+16175555555",
				Text:        make(map[chan int]int),
			},
			wantErr: "json: unsupported type: map[chan int]int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := json.Marshal(tt.message)
			if err == nil {
				t.Error("Expected marshal error but got none")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestMessageValidation(t *testing.T) {
	client := New("test_key")

	tests := []struct {
		name        string
		phoneNumber string
		text        string
		wantErr     string
	}{
		{
			name:        "invalid phone number",
			phoneNumber: "1234567890",
			text:        "test",
			wantErr:     "invalid phone number format",
		},
		{
			name:        "empty phone number",
			phoneNumber: "",
			text:        "test",
			wantErr:     "invalid phone number format",
		},
		{
			name:        "empty text",
			phoneNumber: "+16175555555",
			text:        "",
			wantErr:     "", // Should not error
		},
		{
			name:        "very long text",
			phoneNumber: "+16175555555",
			text:        strings.Repeat("a", 1600),
			wantErr:     "", // API should handle this
		},
		{
			name:        "text with special characters",
			phoneNumber: "+16175555555",
			text:        "Hello 👋 World 🌎!",
			wantErr:     "", // Should handle UTF-8 properly
		},
		{
			name:        "text with newlines and quotes",
			phoneNumber: "+16175555555",
			text:        "Line 1\nLine 2\n\"Quoted text\"",
			wantErr:     "", // Should handle these characters
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First test direct message creation
			msg := MessageRequest{
				PhoneNumber: tt.phoneNumber,
				Text:        tt.text,
			}

			// Test JSON marshaling
			data, err := json.Marshal(msg)
			if err != nil {
				t.Errorf("Failed to marshal message: %v", err)
			}

			// Test JSON unmarshaling
			var unmarshaledMsg MessageRequest
			err = json.Unmarshal(data, &unmarshaledMsg)
			if err != nil {
				t.Errorf("Failed to unmarshal message: %v", err)
			}

			// Verify the message is unchanged
			if unmarshaledMsg.PhoneNumber != tt.phoneNumber {
				t.Errorf("Phone number changed during marshal/unmarshal: got %s, want %s",
					unmarshaledMsg.PhoneNumber, tt.phoneNumber)
			}
			if unmarshaledMsg.Text != tt.text {
				t.Errorf("Text changed during marshal/unmarshal: got %s, want %s",
					unmarshaledMsg.Text, tt.text)
			}

			// Now test the actual Send function
			_, err = client.Send(tt.phoneNumber, tt.text)
			if tt.wantErr != "" {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Expected error containing %q, got %q", tt.wantErr, err.Error())
				}
			} else if err != nil && !strings.Contains(err.Error(), "request failed") {
				// Ignore connection errors, just check for validation errors
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestNetworkErrors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Client
		wantErr string
	}{
		{
			name: "invalid base URL",
			setup: func() *Client {
				c := New("test_key")
				c.baseURL = "http://[::1]:namedport" // Invalid URL
				return c
			},
			wantErr: "failed to create request",
		},
		{
			name: "connection refused",
			setup: func() *Client {
				c := New("test_key")
				c.baseURL = "http://localhost:1"
				c.httpClient.Timeout = 1 * time.Second
				return c
			},
			wantErr: "request failed",
		},
		{
			name: "request timeout",
			setup: func() *Client {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(100 * time.Millisecond)
				}))
				c := New("test_key")
				c.baseURL = ts.URL
				c.httpClient.Timeout = 1 * time.Millisecond
				return c
			},
			wantErr: "timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setup()
			_, err := client.Send("+16175555555", "Test message")
			if err == nil {
				t.Error("Expected error but got none")
			} else if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestResponseErrors(t *testing.T) {
	tests := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
		wantErr string
	}{
		{
			name: "empty response body",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				// Don't write any body
			},
			wantErr: "Invalid JSON",
		},
		{
			name: "invalid json response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{invalid json"))
			},
			wantErr: "Invalid JSON",
		},
		{
			name: "invalid error response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("{invalid error json"))
			},
			wantErr: "Failed to decode error response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer server.Close()

			client := New("test_key")
			client.baseURL = server.URL

			_, err := client.Send("+16175555555", "Test message")
			if err == nil {
				t.Error("Expected error but got none")
			} else if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  APIError
		want string
	}{
		{
			name: "error with type",
			err: APIError{
				Type:    "validation_error",
				Message: "Invalid input",
			},
			want: "textfully: Invalid input (Type: validation_error)",
		},
		{
			name: "error without type",
			err: APIError{
				Message: "Something went wrong",
			},
			want: "textfully: Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("APIError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
