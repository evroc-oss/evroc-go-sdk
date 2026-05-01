// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("manager should not be nil")
	}
	if manager.registry == nil {
		t.Fatal("registry should not be nil")
	}
	if manager.APICallsTotal == nil {
		t.Fatal("APICallsTotal should not be nil")
	}
	if manager.APICallsDuration == nil {
		t.Fatal("APICallsDuration should not be nil")
	}
	if manager.APICallsErrors == nil {
		t.Fatal("APICallsErrors should not be nil")
	}
	if manager.RetriesTotal == nil {
		t.Fatal("RetriesTotal should not be nil")
	}
	if manager.WaiterOperationsTotal == nil {
		t.Fatal("WaiterOperationsTotal should not be nil")
	}
	if manager.AuthTokenRefreshesTotal == nil {
		t.Fatal("AuthTokenRefreshesTotal should not be nil")
	}
}

func TestNewNoOpManager(t *testing.T) {
	manager := NewNoOpManager()
	if manager == nil {
		t.Fatal("manager should not be nil")
	}
	// NoOp manager exists for testing purposes
}

func TestManager_Registry(t *testing.T) {
	manager := NewManager()
	registry := manager.Registry()
	if registry == nil {
		t.Fatal("registry should not be nil")
	}
	if registry != manager.registry {
		t.Error("returned registry should match internal registry")
	}
}

func TestManager_RecordAPICall(t *testing.T) {
	manager := NewManager()
	// Just verify it doesn't panic
	manager.RecordAPICall("GET", "vms", 0.1)
}

func TestManager_RecordAPICallError(t *testing.T) {
	manager := NewManager()
	// Just verify it doesn't panic
	manager.RecordAPICallError("GET", "vms", 0.1, "timeout")
}

func TestManager_RecordRetry(t *testing.T) {
	manager := NewManager()
	// Just verify it doesn't panic
	manager.RecordRetry("GET", "vms", 1.0)
}

func TestManager_RecordWaiterOperation(t *testing.T) {
	manager := NewManager()
	// Just verify it doesn't panic
	manager.RecordWaiterOperation("VirtualMachine", 10.0, 5)
}

func TestManager_RecordWaiterError(t *testing.T) {
	manager := NewManager()
	// Just verify it doesn't panic
	manager.RecordWaiterError("Disk", 60.0, 30)
}

func TestManager_RecordTokenRefresh(t *testing.T) {
	manager := NewManager()
	// Just verify it doesn't panic
	manager.RecordTokenRefresh(0.2)
}

func TestManager_RecordInitialAuth(t *testing.T) {
	manager := NewManager()
	// Just verify it doesn't panic
	manager.RecordInitialAuth(0.5)
}

func TestManager_NilManager(t *testing.T) {
	var manager *Manager
	// All methods should handle nil manager gracefully
	manager.RecordAPICall("GET", "vms", 0.1)
	manager.RecordAPICallError("GET", "vms", 0.1, "error")
	manager.RecordRetry("GET", "vms", 1.0)
	manager.RecordWaiterOperation("VM", 10.0, 5)
	manager.RecordWaiterError("Disk", 60.0, 30)
	manager.RecordTokenRefresh(0.2)
	manager.RecordInitialAuth(0.5)
}

func TestManager_Gather(t *testing.T) {
	manager := NewManager()

	// Record some metrics
	manager.RecordAPICall("GET", "vms", 0.1)
	manager.RecordRetry("POST", "disks", 1.0)

	// Gather metrics
	metricFamilies, err := manager.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	if len(metricFamilies) == 0 {
		t.Error("expected at least some metrics to be gathered")
	}

	// Verify we have our custom metrics (not just Go/process metrics)
	hasCustomMetrics := false
	for _, mf := range metricFamilies {
		if mf.GetName() == "evroc_sdk_api_calls_total" {
			hasCustomMetrics = true
			break
		}
	}

	if !hasCustomMetrics {
		t.Error("expected to find custom SDK metrics")
	}
}
