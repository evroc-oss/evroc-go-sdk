// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package think

import (
	"testing"
	"time"
)

func TestAPIKeyBuilder(t *testing.T) {
	req := NewAPIKeyBuilder("test-key").Build()

	if req.Metadata.Id != "test-key" {
		t.Errorf("Expected id 'test-key', got %s", req.Metadata.Id)
	}
	if req.Spec.ExpiryTimestamp != nil {
		t.Errorf("Expected expiryTimestamp nil, got %v", req.Spec.ExpiryTimestamp)
	}
}

func TestAPIKeyBuilderExpiry(t *testing.T) {
	req := NewAPIKeyBuilder("test-key").
		WithExpiryTimestamp(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).
		Build()

	if req.Metadata.Id != "test-key" {
		t.Errorf("Expected id 'test-key', got %s", req.Metadata.Id)
	}
	if *req.Spec.ExpiryTimestamp != time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC) {
		t.Errorf("Expected expiryTimestamp 2000-01-01, got %v", req.Spec.ExpiryTimestamp)
	}
}

func TestInstanceBuilder(t *testing.T) {
	req := NewInstanceBuilder("test-instance").
		WithModel("test-model").
		Build()

	if req.Metadata.Id != "test-instance" {
		t.Errorf("Expected id 'test-instance', got %s", req.Metadata.Id)
	}
	if req.Spec.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got %s", req.Spec.Model)
	}
	if req.Spec.Size != nil {
		t.Errorf("Expected size nil, got %v", req.Spec.Size)
	}
	if req.Spec.Stopped != nil {
		t.Errorf("Expected stopped nil, got %v", req.Spec.Stopped)
	}
}

func TestInstanceBuilderStoppedSize(t *testing.T) {
	req := NewInstanceBuilder("test-instance").
		WithModel("test-model").
		WithSize("test-size").
		WithStopped(false).
		Build()

	if req.Metadata.Id != "test-instance" {
		t.Errorf("Expected id 'test-instance', got %s", req.Metadata.Id)
	}
	if req.Spec.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got %s", req.Spec.Model)
	}
	if *req.Spec.Size != "test-size" {
		t.Errorf("Expected size 'test-size', got %v", req.Spec.Size)
	}
	if *req.Spec.Stopped {
		t.Errorf("Expected stopped false, got %v", req.Spec.Stopped)
	}
}
