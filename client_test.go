package textfully

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSend(t *testing.T) {
	tests := []struct {
		name       string
		apiKey     string
		phone      string
		text       string
		mockStatus int
		mockResp   interface{}
		wantErr    bool
		errType    string
	}{
		{
			name:       "successful send",
			apiKey:     "test_key",
			phone:      "+16175555555",
			text:       "Test message",
			mockStatus: http.StatusOK,
			mockResp: MessageResponse{
				ID:        "msg_123",
				Status:    "sent",
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name:    "invalid phone number",
			apiKey:  "test_key",
			phone:   "1234567890",
			text:    "Test message",
			wantErr: true,
		},
		{
			name:    "missing api key",
			apiKey:  "",
			phone:   "+16175555555",
			text:    "Test message",
			wantErr: true,
		},
		{
			name:       "authentication error",
			apiKey:     "invalid_key",
			phone:      "+16175555555",
			text:       "Test message",
			mockStatus: http.StatusUnauthorized,
			mockResp: map[string]interface{}{
				"error": map[string]string{
					"type":    "authentication_error",
					"message": "Invalid API key",
				},
			},
			wantErr: true,
			errType: "authentication_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check method
				assert.Equal(t, http.MethodPost, r.Method)

				// Check auth header if api key provided
				if tt.apiKey != "" {
					assert.Equal(t, "Bearer "+tt.apiKey, r.Header.Get("Authorization"))
				}

				// Send response
				w.WriteHeader(tt.mockStatus)
				if tt.mockResp != nil {
					json.NewEncoder(w).Encode(tt.mockResp)
				}
			}))
			defer server.Close()

			client := New(tt.apiKey)
			client.baseURL = server.URL

			resp, err := client.Send(tt.phone, tt.text)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != "" {
					apiErr, ok := err.(*APIError)
					assert.True(t, ok)
					assert.Equal(t, tt.errType, apiErr.Type)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "msg_123", resp.ID)
			}
		})
	}
}
