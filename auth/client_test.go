// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/config"
)

// mockTokenResponse represents an OAuth2 token response.
type mockTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// createMockTokenServer creates a test OAuth2 token server.
func createMockTokenServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a POST request
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Verify grant type
		if r.FormValue("grant_type") != "password" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "unsupported_grant_type",
			})
			return
		}

		// Verify credentials
		username := r.FormValue("username")
		password := r.FormValue("password")
		if username != "test@example.com" || password != "testpass" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid_grant",
			})
			return
		}

		// Return successful token response
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockTokenResponse{
			AccessToken: "test-access-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		})
	}))
}

func TestNewClient(t *testing.T) {
	server := createMockTokenServer()
	defer server.Close()

	tests := []struct {
		name    string
		cfg     config.AuthConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: config.AuthConfig{
				TokenURL: server.URL,
				ClientID: "test-client",
				Username: "test@example.com",
				Password: "testpass",
			},
			wantErr: false,
		},
		{
			name: "missing username",
			cfg: config.AuthConfig{
				TokenURL: server.URL,
				ClientID: "test-client",
				Password: "testpass",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			cfg: config.AuthConfig{
				TokenURL: server.URL,
				ClientID: "test-client",
				Username: "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "missing token url",
			cfg: config.AuthConfig{
				ClientID: "test-client",
				Username: "test@example.com",
				Password: "testpass",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(context.Background(), tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestClient_Token(t *testing.T) {
	server := createMockTokenServer()
	defer server.Close()

	client, err := NewClient(context.Background(), config.AuthConfig{
		TokenURL: server.URL,
		ClientID: "test-client",
		Username: "test@example.com",
		Password: "testpass",
	})
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	token, err := client.Token(context.Background())
	if err != nil {
		t.Fatalf("Token() failed: %v", err)
	}

	if token != "test-access-token" {
		t.Errorf("Token() = %v, want test-access-token", token)
	}
}

func TestClient_AddAuthHeader(t *testing.T) {
	server := createMockTokenServer()
	defer server.Close()

	client, err := NewClient(context.Background(), config.AuthConfig{
		TokenURL: server.URL,
		ClientID: "test-client",
		Username: "test@example.com",
		Password: "testpass",
	})
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com", nil)
	err = client.AddAuthHeader(context.Background(), req)
	if err != nil {
		t.Fatalf("AddAuthHeader() failed: %v", err)
	}

	authHeader := req.Header.Get("Authorization")
	if authHeader != "Bearer test-access-token" {
		t.Errorf("Authorization header = %v, want Bearer test-access-token", authHeader)
	}
}

func TestClient_IsTokenValid(t *testing.T) {
	server := createMockTokenServer()
	defer server.Close()

	client, err := NewClient(context.Background(), config.AuthConfig{
		TokenURL: server.URL,
		ClientID: "test-client",
		Username: "test@example.com",
		Password: "testpass",
	})
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	// Token should be valid immediately after creation
	if !client.IsTokenValid() {
		t.Error("IsTokenValid() = false, want true for fresh token")
	}

	// Manually set an expired token
	client.tokenMu.Lock()
	client.token.Expiry = time.Now().Add(-1 * time.Hour)
	client.tokenMu.Unlock()

	if client.IsTokenValid() {
		t.Error("IsTokenValid() = true, want false for expired token")
	}
}

func TestClient_HTTPClient(t *testing.T) {
	server := createMockTokenServer()
	defer server.Close()

	client, err := NewClient(context.Background(), config.AuthConfig{
		TokenURL: server.URL,
		ClientID: "test-client",
		Username: "test@example.com",
		Password: "testpass",
	})
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	httpClient := client.HTTPClient()
	if httpClient == nil {
		t.Error("HTTPClient() returned nil")
	}
}
