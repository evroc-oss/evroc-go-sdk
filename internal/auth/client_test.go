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

func TestClient_Options(t *testing.T) {
	server := createMockTokenServer()
	defer server.Close()

	ctx := context.Background()
	customHTTP := &http.Client{Timeout: 5 * time.Second}
	client, err := NewClient(ctx, config.AuthConfig{
		TokenURL: server.URL,
		ClientID: "test-client",
		Username: "test@example.com",
		Password: "testpass",
	}, WithHTTPClient(customHTTP), WithMetrics(nil))
	if err != nil {
		t.Fatalf("NewClient with options failed: %v", err)
	}
	if httpClient, err := client.GetHTTPClient(ctx); err != nil || httpClient == nil {
		t.Error("GetHTTPClient() failed")
	}

	token, err := client.Token(ctx)
	if err != nil || token == "" {
		t.Error("Token() failed")
	}

	if err := client.ForceRefresh(ctx); err != nil {
		t.Errorf("ForceRefresh() failed: %v", err)
	}
}

func TestInteractive(t *testing.T) {
	cfg := OIDCConfig{
		ClientID: "test-client",
		TokenURL: "https://auth.example.com/token",
		AuthURL:  "https://auth.example.com/auth",
	}
	url := GetAuthorizationURL(cfg, "verifier", "state")
	if url == "" {
		t.Error("GetAuthorizationURL returned empty string")
	}

	if GenerateVerifier() == "" {
		t.Error("GenerateVerifier returned empty string")
	}
	if GenerateState() == "" {
		t.Error("GenerateState returned empty string")
	}
}

// mockMetrics implements AuthMetricsRecorder for testing.
type mockMetrics struct {
	refreshCalls      int
	refreshErrorCalls int
	authCalls         int
	authErrorCalls    int
	lastErrorType     string
}

func (m *mockMetrics) RecordTokenRefresh(duration float64)                    { m.refreshCalls++ }
func (m *mockMetrics) RecordTokenRefreshError(duration float64, errType string) { m.refreshErrorCalls++; m.lastErrorType = errType }
func (m *mockMetrics) RecordInitialAuth(duration float64)                     { m.authCalls++ }
func (m *mockMetrics) RecordInitialAuthError(duration float64, errType string) { m.authErrorCalls++; m.lastErrorType = errType }

func TestMetricsTokenSource(t *testing.T) {
	server := createMockTokenServer()
	defer server.Close()

	metrics := &mockMetrics{}
	ctx := context.Background()

	client, err := NewClient(ctx, config.AuthConfig{
		TokenURL: server.URL,
		ClientID: "test-client",
		Username: "test@example.com",
		Password: "testpass",
	}, WithMetrics(metrics))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// Initial auth should record metrics
	if metrics.authCalls != 1 {
		t.Errorf("Expected 1 auth call, got %d", metrics.authCalls)
	}

	// Get token through token source
	_, err = client.Token(ctx)
	if err != nil {
		t.Fatalf("Token() failed: %v", err)
	}

	// Test with token-based client (direct token path with metrics)
	validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk5OTk5OTk5OTl9.signature"
	metrics2 := &mockMetrics{}
	client2, err := NewClient(ctx, config.AuthConfig{
		TokenURL: server.URL,
		ClientID: "test-client",
		Token:    validToken,
	}, WithMetrics(metrics2))
	if err != nil {
		t.Fatalf("NewClient with token failed: %v", err)
	}
	_, err = client2.Token(ctx)
	if err != nil {
		t.Fatalf("Token() with metrics failed: %v", err)
	}
}

func TestClassifyAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"nil error", nil, "none"},
		{"invalid_grant", &mockError{"oauth2: invalid_grant"}, "invalid_credentials"},
		{"invalid_client", &mockError{"oauth2: invalid_client"}, "invalid_client"},
		{"unauthorized_client", &mockError{"oauth2: unauthorized_client"}, "unauthorized_client"},
		{"unsupported_grant_type", &mockError{"oauth2: unsupported_grant_type"}, "unsupported_grant"},
		{"invalid_scope", &mockError{"oauth2: invalid_scope"}, "invalid_scope"},
		{"timeout", &mockError{"context deadline exceeded"}, "timeout"},
		{"timeout2", &mockError{"timeout occurred"}, "timeout"},
		{"network_error", &mockError{"connection refused"}, "network_error"},
		{"network_error2", &mockError{"no such host"}, "network_error"},
		{"unknown", &mockError{"some other error"}, "unknown_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyAuthError(tt.err)
			if result != tt.expected {
				t.Errorf("classifyAuthError(%v) = %s, want %s", tt.err, result, tt.expected)
			}
		})
	}
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func TestExtractTokenExpiry(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		wantValid bool
	}{
		{"invalid format", "not-a-jwt", false},
		{"two parts only", "header.payload", false},
		{"invalid base64", "header.!!!invalid!!!.signature", false},
		{"invalid json", "header.aW52YWxpZGpzb24.signature", false},
		{"no expiry", "eyJhbGciOiJIUzI1NiJ9.e30.signature", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expiry := extractTokenExpiry(tt.token)
			now := time.Now()
			// All invalid tokens should return current time or earlier
			if tt.wantValid {
				if expiry.Before(now.Add(-1 * time.Second)) {
					t.Error("Expected future expiry for valid token")
				}
			} else {
				if expiry.After(now.Add(1 * time.Second)) {
					t.Error("Expected past/current expiry for invalid token")
				}
			}
		})
	}
}

func TestNewClientWithToken(t *testing.T) {
	server := createMockTokenServer()
	defer server.Close()

	ctx := context.Background()

	// Test with direct token
	validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk5OTk5OTk5OTl9.signature"
	client, err := NewClient(ctx, config.AuthConfig{
		TokenURL: server.URL,
		ClientID: "test-client",
		Token:    validToken,
	})
	if err != nil {
		t.Fatalf("NewClient with token failed: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client, got nil")
	}

	// Test with token and refresh token
	client, err = NewClient(ctx, config.AuthConfig{
		TokenURL:     server.URL,
		ClientID:     "test-client",
		Token:        validToken,
		RefreshToken: "refresh-token",
	})
	if err != nil {
		t.Fatalf("NewClient with token and refresh failed: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client, got nil")
	}
}

func TestWithHTTPClientNil(t *testing.T) {
	err := WithHTTPClient(nil)(&Client{})
	if err == nil {
		t.Error("WithHTTPClient(nil) should return error")
	}
}

func TestIsTokenValidNilToken(t *testing.T) {
	client := &Client{token: nil}
	if client.IsTokenValid() {
		t.Error("IsTokenValid() should return false for nil token")
	}
}

func TestNewClientFromManualCallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.FormValue("grant_type") == "authorization_code" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockTokenResponse{
				AccessToken: "test-access-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			})
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	cfg := OIDCConfig{
		ClientID: "test-client",
		TokenURL: server.URL,
		AuthURL:  "https://auth.example.com/auth",
	}

	tests := []struct {
		name        string
		callbackURL string
		state       string
		wantErr     bool
	}{
		{"valid callback", "http://localhost:8000?code=test-code&state=test-state", "test-state", false},
		{"invalid url", "://invalid", "test-state", true},
		{"oauth error", "http://localhost:8000?error=access_denied&error_description=User denied", "test-state", true},
		{"oauth error no desc", "http://localhost:8000?error=access_denied", "test-state", true},
		{"no state", "http://localhost:8000?code=test-code", "test-state", true},
		{"state mismatch", "http://localhost:8000?code=test-code&state=wrong-state", "test-state", true},
		{"no code", "http://localhost:8000?state=test-state", "test-state", true},
		{"missing config", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCfg := cfg
			if tt.name == "missing config" {
				testCfg = OIDCConfig{}
			}
			_, _, err := NewClientFromManualCallback(context.Background(), testCfg, tt.callbackURL, "verifier", tt.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClientFromManualCallback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthenticateError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid_grant",
		})
	}))
	defer server.Close()

	metrics := &mockMetrics{}
	_, err := NewClient(context.Background(), config.AuthConfig{
		TokenURL: server.URL,
		ClientID: "test-client",
		Username: "wrong@example.com",
		Password: "wrongpass",
	}, WithMetrics(metrics))

	if err == nil {
		t.Error("Expected error for invalid credentials")
	}
	if metrics.authErrorCalls != 1 {
		t.Errorf("Expected 1 auth error call, got %d", metrics.authErrorCalls)
	}
}

func TestGetHTTPClientError(t *testing.T) {
	client := &Client{tokenSource: nil}
	_, err := client.GetHTTPClient(context.Background())
	if err == nil {
		t.Error("GetHTTPClient() should error with nil token source")
	}
}

func TestTokenError(t *testing.T) {
	client := &Client{tokenSource: nil}
	_, err := client.Token(context.Background())
	if err == nil {
		t.Error("Token() should error with nil token source")
	}
}

func TestAddAuthHeaderError(t *testing.T) {
	client := &Client{tokenSource: nil}
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com", nil)
	err := client.AddAuthHeader(context.Background(), req)
	if err == nil {
		t.Error("AddAuthHeader() should error with nil token source")
	}
}

func TestNewClientMissingClientID(t *testing.T) {
	_, err := NewClient(context.Background(), config.AuthConfig{
		TokenURL: "https://token.example.com",
		Username: "test@example.com",
		Password: "testpass",
	})
	if err == nil {
		t.Error("NewClient should error with missing ClientID")
	}
}

func TestNewClientWithScopes(t *testing.T) {
	server := createMockTokenServer()
	defer server.Close()

	// Test with custom scopes that already include offline_access
	client, err := NewClient(context.Background(), config.AuthConfig{
		TokenURL: server.URL,
		ClientID: "test-client",
		Username: "test@example.com",
		Password: "testpass",
		Scopes:   []string{"openid", "offline_access", "profile"},
	})
	if err != nil {
		t.Fatalf("NewClient with scopes failed: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client, got nil")
	}

	// Test with custom scopes without offline_access (should be added)
	client, err = NewClient(context.Background(), config.AuthConfig{
		TokenURL: server.URL,
		ClientID: "test-client",
		Username: "test@example.com",
		Password: "testpass",
		Scopes:   []string{"openid", "profile"},
	})
	if err != nil {
		t.Fatalf("NewClient with scopes (no offline_access) failed: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client, got nil")
	}
}
