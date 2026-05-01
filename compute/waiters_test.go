// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

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
	ready := `{"apiVersion":"` + builderAPIVersion + `","kind":"Test","metadata":{"id":"test"},"status":{"conditions":[{"type":"Ready","status":"True"}]}}`
	notReady := `{"apiVersion":"` + builderAPIVersion + `","kind":"Test","metadata":{"id":"test"},"status":{"conditions":[{"type":"Ready","status":"False"}]}}`

	t.Run("VM_WaitForReady", func(t *testing.T) {
		client, cleanup := setupWaiter(t, ready, http.StatusOK)
		defer cleanup()
		vm, err := client.VirtualMachines().WaitForReady(ctx, "test", 2*time.Second, WithPollingInterval(100*time.Millisecond), WithProgressCallback(func(int, time.Duration) {}))
		if err != nil {
			t.Errorf("Failed: %v", err)
		}
		if vm == nil {
			t.Error("Expected non-nil VM")
		}
		if vm != nil && (vm.Metadata.Id != "test") {
			t.Errorf("Expected VM ID 'test', got %v", vm.Metadata.Id)
		}
	})

	t.Run("VM_WaitForReady_Timeout", func(t *testing.T) {
		client, cleanup := setupWaiter(t, notReady, http.StatusOK)
		defer cleanup()
		if _, err := client.VirtualMachines().WaitForReady(ctx, "test", 500*time.Millisecond, WithPollingInterval(100*time.Millisecond)); err == nil {
			t.Error("Should timeout")
		}
	})

	t.Run("VM_WaitForDeleted", func(t *testing.T) {
		client, cleanup := setupWaiter(t, "", http.StatusNotFound)
		defer cleanup()
		if err := client.VirtualMachines().WaitForDeleted(ctx, "test", 2*time.Second, WithPollingInterval(100*time.Millisecond)); err != nil {
			t.Errorf("Failed: %v", err)
		}
	})

	t.Run("Disk_WaitForReady", func(t *testing.T) {
		client, cleanup := setupWaiter(t, ready, http.StatusOK)
		defer cleanup()
		disk, err := client.Disks().WaitForReady(ctx, "test", 2*time.Second, WithPollingInterval(100*time.Millisecond))
		if err != nil {
			t.Errorf("Failed: %v", err)
		}
		if disk == nil {
			t.Error("Expected non-nil disk")
		}
		if disk != nil && (disk.Metadata.Id != "test") {
			t.Errorf("Expected disk ID 'test', got %v", disk.Metadata.Id)
		}
	})

	t.Run("Disk_WaitForDeleted", func(t *testing.T) {
		client, cleanup := setupWaiter(t, "", http.StatusNotFound)
		defer cleanup()
		if err := client.Disks().WaitForDeleted(ctx, "test", 2*time.Second, WithPollingInterval(100*time.Millisecond)); err != nil {
			t.Errorf("Failed: %v", err)
		}
	})

	t.Run("HotswapDiskAttachment_WaitForReady", func(t *testing.T) {
		client, cleanup := setupWaiter(t, ready, http.StatusOK)
		defer cleanup()
		attachment, err := client.HotswapDiskAttachments().WaitForReady(ctx, "test", 2*time.Second, WithPollingInterval(100*time.Millisecond), WithExponentialBackoff(100*time.Millisecond, 1*time.Second, 2.0))
		if err != nil {
			t.Errorf("Failed: %v", err)
		}
		if attachment == nil {
			t.Error("Expected non-nil attachment")
		}
		if attachment != nil && (attachment.Metadata.Id != "test") {
			t.Errorf("Expected attachment ID 'test', got %v", attachment.Metadata.Id)
		}
	})

	t.Run("PlacementGroup_WaitForReady", func(t *testing.T) {
		client, cleanup := setupWaiter(t, ready, http.StatusOK)
		defer cleanup()
		pg, err := client.PlacementGroups().WaitForReady(ctx, "test", 2*time.Second, WithPollingInterval(100*time.Millisecond))
		if err != nil {
			t.Errorf("Failed: %v", err)
		}
		if pg == nil {
			t.Error("Expected non-nil placement group")
		}
		if pg != nil && (pg.Metadata.Id != "test") {
			t.Errorf("Expected placement group ID 'test', got %v", pg.Metadata.Id)
		}
	})

	t.Run("PlacementGroup_WaitForReady_Timeout", func(t *testing.T) {
		client, cleanup := setupWaiter(t, notReady, http.StatusOK)
		defer cleanup()
		if _, err := client.PlacementGroups().WaitForReady(ctx, "test", 500*time.Millisecond, WithPollingInterval(100*time.Millisecond)); err == nil {
			t.Error("Should timeout")
		}
	})
}
