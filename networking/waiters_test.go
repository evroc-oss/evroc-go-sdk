// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
)

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
	ctx := context.Background()
	ready := `{"apiVersion":"v1beta1","kind":"Test","metadata":{"id":"test"},"status":{"conditions":[{"type":"Ready","status":"True"}]}}`
	notReady := `{"apiVersion":"v1beta1","kind":"Test","metadata":{"id":"test"},"status":{"conditions":[{"type":"Ready","status":"False"}]}}`

	t.Run("PublicIP_WaitForReady", func(t *testing.T) {
		client, cleanup := setupWaiter(t, ready, http.StatusOK)
		defer cleanup()
		ip, err := client.PublicIPs().WaitForReady(ctx, "test", 2*time.Second, WithPollingInterval(100*time.Millisecond), WithProgressCallback(func(int, time.Duration) {}))
		if err != nil {
			t.Errorf("Failed: %v", err)
		}
		if ip == nil {
			t.Error("Expected non-nil public IP")
		}
		if ip != nil && (ip.Metadata.Id != "test") {
			t.Errorf("Expected public IP ID 'test', got %v", ip.Metadata.Id)
		}
	})

	t.Run("PublicIP_WaitForReady_Timeout", func(t *testing.T) {
		client, cleanup := setupWaiter(t, notReady, http.StatusOK)
		defer cleanup()
		if _, err := client.PublicIPs().WaitForReady(ctx, "test", 500*time.Millisecond, WithPollingInterval(100*time.Millisecond)); err == nil {
			t.Error("Should timeout")
		}
	})

	t.Run("SecurityGroup_WaitForReady", func(t *testing.T) {
		client, cleanup := setupWaiter(t, ready, http.StatusOK)
		defer cleanup()
		sg, err := client.SecurityGroups().WaitForReady(ctx, "test", 2*time.Second, WithPollingInterval(100*time.Millisecond))
		if err != nil {
			t.Errorf("Failed: %v", err)
		}
		if sg == nil {
			t.Error("Expected non-nil security group")
		}
		if sg != nil && (sg.Metadata.Id != "test") {
			t.Errorf("Expected security group ID 'test', got %v", sg.Metadata.Id)
		}
	})

	t.Run("SecurityGroup_WaitForReady_WithExponentialBackoff", func(t *testing.T) {
		client, cleanup := setupWaiter(t, ready, http.StatusOK)
		defer cleanup()
		sg, err := client.SecurityGroups().WaitForReady(ctx, "test", 2*time.Second, WithExponentialBackoff(100*time.Millisecond, 1*time.Second, 2.0))
		if err != nil {
			t.Errorf("Failed: %v", err)
		}
		if sg == nil {
			t.Error("Expected non-nil security group")
		}
		if sg != nil && (sg.Metadata.Id != "test") {
			t.Errorf("Expected security group ID 'test', got %v", sg.Metadata.Id)
		}
	})
}
