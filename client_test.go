package textfully

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
			name:    "invalid phone number",
			apiKey:  "test_key",
			phone:   "1234567890",
			text:    "Test message",
			wantErr: true,
			setupMock: func(w http.ResponseWriter) {
				// No mock needed - validation happens before request
			},
		},
		{
			name:    "missing api key",
			apiKey:  "",
			phone:   "+16175555555",
			text:    "Test message",
			wantErr: true,
			setupMock: func(w http.ResponseWriter) {
				// No mock needed - validation happens before request
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check method
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				// Check auth header if api key provided
				if tt.apiKey != "" {
					if auth := r.Header.Get("Authorization"); auth != "Bearer "+tt.apiKey {
						t.Errorf("Expected Authorization header 'Bearer %s', got '%s'", tt.apiKey, auth)
					}
				}

				// Set up mock response
				tt.setupMock(w)
			}))
			defer server.Close()

			client := New(tt.apiKey)
			client.baseURL = server.URL

			resp, err := client.Send(tt.phone, tt.text)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
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
