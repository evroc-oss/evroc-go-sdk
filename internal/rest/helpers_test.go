// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestRESTMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"apiVersion":"v1","kind":"Test","metadata":{"id":"test"}}`))
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"apiVersion":"v1","kind":"Test","metadata":{"id":"test"}}`))
		case http.MethodPatch:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"apiVersion":"v1","kind":"Test","metadata":{"id":"test"}}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()

	// Test Get
	if _, err := client.Get(ctx, "/test", nil); err != nil {
		t.Errorf("Get failed: %v", err)
	}

	// Test Patch
	if _, err := client.Patch(ctx, "/test", map[string]string{"key": "value"}); err != nil {
		t.Errorf("Patch failed: %v", err)
	}

	// Test Delete
	if _, err := client.Delete(ctx, "/test"); err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

type testResource struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   map[string]string `json:"metadata"`
}

func TestResourceHelpers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"apiVersion":"v1","kind":"Test","metadata":{"id":"test"}}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	ctx := context.Background()

	// Test GetResource
	if _, err := GetResource[*testResource](ctx, client, "/test"); err != nil {
		t.Errorf("GetResource failed: %v", err)
	}

	// Test ListResources
	type testList struct {
		Items []testResource `json:"items"`
	}
	if _, err := ListResources[*testList](ctx, client, "/tests"); err != nil {
		t.Errorf("ListResources failed: %v", err)
	}

	// Test ListResourcesWithQuery
	query := url.Values{}
	query.Set("filter", "test")
	if _, err := ListResourcesWithQuery[*testList](ctx, client, "/tests", query); err != nil {
		t.Errorf("ListResourcesWithQuery failed: %v", err)
	}

	// Test CreateResource
	req := &testResource{APIVersion: "v1", Kind: "Test"}
	if _, err := CreateResource[*testResource](ctx, client, "/tests", req); err != nil {
		t.Errorf("CreateResource failed: %v", err)
	}

	// Test PatchResource
	if _, err := PatchResource[*testResource](ctx, client, "/test", req); err != nil {
		t.Errorf("PatchResource failed: %v", err)
	}

	// Test DeleteResource
	if err := DeleteResource(ctx, client, "/test"); err != nil {
		t.Errorf("DeleteResource failed: %v", err)
	}
}

func TestServicePath(t *testing.T) {
	path := NewServicePath("test", "v1")
	tests := []struct {
		name string
		fn   func() string
	}{
		{"CollectionPath", func() string { return path.CollectionPath("proj", "region", "resource") }},
		{"ResourcePath", func() string { return path.ResourcePath("proj", "region", "resource", "name") }},
		{"ProjectCollectionPath", func() string { return path.ProjectCollectionPath("proj", "resource") }},
		{"ProjectResourcePath", func() string { return path.ProjectResourcePath("proj", "resource", "name") }},
		{"GlobalCollectionPath", func() string { return path.GlobalCollectionPath("resource") }},
		{"GlobalResourcePath", func() string { return path.GlobalResourcePath("resource", "name") }},
	}
	for _, tt := range tests {
		if p := tt.fn(); p == "" {
			t.Errorf("%s returned empty", tt.name)
		}
	}
}

func TestFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	ctx := context.Background()

	type testList struct {
		Items []testResource `json:"items"`
	}

	labelFilter := WithLabels(map[string]string{"app": "test", "env": "prod"})
	if _, err := ListWithFilters[*testList](ctx, client, "/tests", labelFilter); err != nil {
		t.Errorf("ListWithFilters failed: %v", err)
	}
}
