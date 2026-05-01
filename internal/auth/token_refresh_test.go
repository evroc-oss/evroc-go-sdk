// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

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

// TestTokenRefreshWorks verifies that the SDK properly refreshes expired tokens
func TestTokenRefreshWorks(t *testing.T) {
	var passwordAuthCalls int
	var refreshTokenCalls int

	// Create mock OAuth2 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		grantType := r.FormValue("grant_type")

		switch grantType {
		case "password":
			passwordAuthCalls++
			t.Logf("OAuth2 server: password auth (call #%d)", passwordAuthCalls)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "initial-token",
				"refresh_token": "valid-refresh-token",
				"token_type":    "Bearer",
				"expires_in":    1, // Expire in 1 second
			})
		case "refresh_token":
			refreshTokenCalls++
			t.Logf("OAuth2 server: token refresh (call #%d)", refreshTokenCalls)

			refreshToken := r.FormValue("refresh_token")
			if refreshToken != "valid-refresh-token" {
				t.Errorf("Wrong refresh token: %s", refreshToken)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "refreshed-token",
				"refresh_token": "valid-refresh-token",
				"token_type":    "Bearer",
				"expires_in":    3600,
			})
		default:
			t.Logf("OAuth2 server: unknown grant_type=%s", grantType)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	// Create SDK auth client with password
	ctx := context.Background()
	client, err := NewClient(ctx, config.AuthConfig{
		TokenURL: server.URL,
		ClientID: "test-client",
		Username: "test@example.com",
		Password: "testpass",
	})
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	t.Logf("Step 1: Initial authentication")
	if passwordAuthCalls != 1 {
		t.Errorf("Expected 1 password auth call after init, got %d", passwordAuthCalls)
	}

	// Wait for token to expire
	t.Logf("Step 2: Waiting 2 seconds for token to expire...")
	time.Sleep(2 * time.Second)

	// Request a new token - this should trigger refresh
	t.Logf("Step 3: Getting token after expiry (should trigger refresh)")
	token, err := client.Token()
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}
	t.Logf("Got token: %s", token)

	// Verify results
	t.Logf("\nResults:")
	t.Logf("  Password auth calls: %d", passwordAuthCalls)
	t.Logf("  Refresh token calls: %d", refreshTokenCalls)

	if refreshTokenCalls == 0 {
		t.Errorf("BUG: Token refresh was never called!")
		t.Errorf("Expected: 1 password auth + 1 refresh")
		t.Errorf("Actual: %d password auth + %d refresh", passwordAuthCalls, refreshTokenCalls)
	}

	if passwordAuthCalls > 1 {
		t.Errorf("BUG: SDK re-authenticated with password instead of using refresh token")
		t.Errorf("Password was called %d times (should be 1)", passwordAuthCalls)
	}

	if refreshTokenCalls == 1 && passwordAuthCalls == 1 {
		t.Logf("✓ SUCCESS: Token refresh works correctly!")
	}

	if token != "refreshed-token" {
		t.Errorf("Expected refreshed token, got: %s", token)
	}
}

// TestTokenRefreshWithTokenOnly verifies refresh works in token-only mode
func TestTokenRefreshWithTokenOnly(t *testing.T) {
	var refreshTokenCalls int

	// Create mock OAuth2 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		grantType := r.FormValue("grant_type")

		if grantType == "refresh_token" {
			refreshTokenCalls++
			t.Logf("OAuth2 server: token refresh (call #%d)", refreshTokenCalls)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "new-refreshed-token",
				"refresh_token": "still-valid-refresh-token",
				"token_type":    "Bearer",
				"expires_in":    3600,
			})
		} else {
			t.Errorf("Unexpected grant_type: %s (expected only refresh_token)", grantType)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	// Create expired access token
	expiredToken := createExpiredJWT()

	// Create SDK auth client with tokens only (no password)
	ctx := context.Background()
	client, err := NewClient(ctx, config.AuthConfig{
		TokenURL:     server.URL,
		ClientID:     "test-client",
		Token:        expiredToken,
		RefreshToken: "still-valid-refresh-token",
	})
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	t.Logf("Step 1: Client created with expired token + refresh token")

	// Request a token - should refresh immediately since it's expired
	t.Logf("Step 2: Getting token (should trigger refresh)")
	token, err := client.Token()
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	t.Logf("Got token: %s", token)
	t.Logf("Refresh token calls: %d", refreshTokenCalls)

	if refreshTokenCalls != 1 {
		t.Errorf("Expected exactly 1 refresh call, got %d", refreshTokenCalls)
	}

	if token != "new-refreshed-token" {
		t.Errorf("Expected new-refreshed-token, got: %s", token)
	}

	t.Logf("✓ SUCCESS: Token-only mode refresh works!")
}

// createExpiredJWT creates a JWT token that's already expired
func createExpiredJWT() string {
	// This is a JWT with exp set to a time in the past
	// Header: {"alg":"none","typ":"JWT"}
	// Payload: {"exp":1000000000} (Sept 2001 - definitely expired)
	// No signature (algorithm "none")
	return "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJleHAiOjEwMDAwMDAwMDB9."
}
