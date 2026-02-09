// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package rest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evroc-oss/evroc-go-sdk/config"
	"github.com/evroc-oss/evroc-go-sdk/internal/auth"
)

// createMockAuthClient creates a mock auth client for testing.
func createMockAuthClient() *auth.Client {
	// Create a mock OAuth2 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`))
	}))

	// Create auth client with mock server
	client, _ := auth.NewClient(context.Background(), config.AuthConfig{
		TokenURL: server.URL,
		ClientID: "test",
		Username: "test@example.com",
		Password: "test",
	})
	return client
}

func TestClient_Get(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodGet {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("Authorization = %v, want Bearer test-token", auth)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Create auth client and get auto-auth HTTP client
	authClient := createMockAuthClient()
	httpClient, err := authClient.GetHTTPClient(context.Background())
	if err != nil {
		t.Fatalf("Failed to get HTTP client: %v", err)
	}

	// Create REST client
	client, err := NewClient(Config{
		BaseURL:    server.URL,
		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	// Test GET request
	resp, err := client.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
	}
}

func TestClient_Post(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %v, want application/json", ct)
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

	// Create auth client and get auto-auth HTTP client
	authClient := createMockAuthClient()
	httpClient, err := authClient.GetHTTPClient(context.Background())
	if err != nil {
		t.Fatalf("Failed to get HTTP client: %v", err)
	}

	// Create REST client
	client, _ := NewClient(Config{
		BaseURL:    server.URL,
		HTTPClient: httpClient,
	})

	body := map[string]string{"name": "test"}
	resp, err := client.Post(context.Background(), "/test", body)
	if err != nil {
		t.Fatalf("Post() failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %v, want 201", resp.StatusCode)
	}
}

func TestAPIError(t *testing.T) {
	// Test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"reason":"Resource not found"}`))
	}))
	defer server.Close()

	// Create auth client and get auto-auth HTTP client
	authClient := createMockAuthClient()
	httpClient, err := authClient.GetHTTPClient(context.Background())
	if err != nil {
		t.Fatalf("Failed to get HTTP client: %v", err)
	}

	// Create REST client
	client, _ := NewClient(Config{
		BaseURL:    server.URL,
		HTTPClient: httpClient,
	})

	_, err = client.Get(context.Background(), "/notfound", nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	apiErr := &APIError{}
	ok := errors.As(err, &apiErr)
	if !ok {
		t.Fatalf("Error type = %T, want *APIError", err)
	}

	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %v, want 404", apiErr.StatusCode)
	}
	if apiErr.Reason != "Resource not found" {
		t.Errorf("Reason = %v, want Resource not found", apiErr.Reason)
	}
}

func TestDecodeJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"test","count":42}`))
	}))
	defer server.Close()

	// Create auth client and get auto-auth HTTP client
	authClient := createMockAuthClient()
	httpClient, err := authClient.GetHTTPClient(context.Background())
	if err != nil {
		t.Fatalf("Failed to get HTTP client: %v", err)
	}

	// Create REST client
	client, _ := NewClient(Config{
		BaseURL:    server.URL,
		HTTPClient: httpClient,
	})

	resp, _ := client.Get(context.Background(), "/test", nil)

	var result struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	err = DecodeJSON(resp, &result)
	if err != nil {
		t.Fatalf("DecodeJSON() failed: %v", err)
	}

	if result.Name != "test" {
		t.Errorf("Name = %v, want test", result.Name)
	}
	if result.Count != 42 {
		t.Errorf("Count = %v, want 42", result.Count)
	}
}
