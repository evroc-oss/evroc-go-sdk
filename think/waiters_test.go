// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package think

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
)

type mockContextProvider struct{}

func (m *mockContextProvider) DefaultProject() string      { return "test-project" }
func (m *mockContextProvider) DefaultRegion() string       { return "test-region" }
func (m *mockContextProvider) DefaultOrganization() string { return "test-org" }

func setupWaiter(t *testing.T, response string, statusCode int) (*Client, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if response != "" {
			w.Write([]byte(response))
		}
	}))
	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	return NewClient(restClient, &mockContextProvider{}), server.Close
}

func TestWaiters(t *testing.T) {
	created := `{"apiVersion":"` + builderAPIVersion + `","kind":"Test","metadata":{"id":"test"},"status":{"phase":"Created"}}`
	running := `{"apiVersion":"` + builderAPIVersion + `","kind":"Test","metadata":{"id":"test"},"status":{"phase":"Running"}}`
	stopped := `{"apiVersion":"` + builderAPIVersion + `","kind":"Test","metadata":{"id":"test"},"status":{"phase":"Stopped"}}`

	t.Run("Instance_WaitForRunning", func(t *testing.T) {
		client, cleanup := setupWaiter(t, running, http.StatusOK)
		defer cleanup()
		instance, err := client.Instances().WaitForReady(t.Context(), "test", time.Second, WithPollingInterval(100*time.Millisecond))
		if err != nil {
			t.Errorf("Failed: %v", err)
		}
		if instance == nil {
			t.Errorf("Expected non-nil instance")
		}
		if instance != nil && instance.Metadata.Id != "test" {
			t.Errorf("Expected instance ID 'test', got %v", instance.Metadata.Id)
		}
	})

	t.Run("Instance_WaitForStopped", func(t *testing.T) {
		client, cleanup := setupWaiter(t, stopped, http.StatusOK)
		defer cleanup()
		instance, err := client.Instances().WaitForStopped(t.Context(), "test", time.Second, WithPollingInterval(100*time.Millisecond))
		if err != nil {
			t.Errorf("Failed: %v", err)
		}
		if instance == nil {
			t.Errorf("Expected non-nil instance")
		}
		if instance != nil && instance.Metadata.Id != "test" {
			t.Errorf("Expected instance ID 'test', got %v", instance.Metadata.Id)
		}
	})

	t.Run("Instance_Timeout", func(t *testing.T) {
		client, cleanup := setupWaiter(t, created, http.StatusOK)
		defer cleanup()
		_, err := client.Instances().WaitForStopped(t.Context(), "test", time.Second, WithPollingInterval(100*time.Millisecond))
		if err == nil {
			t.Error("Should timeout")
		}
	})
}
