// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/types/storage"
)

type mockContextProvider struct{}

func (m *mockContextProvider) DefaultProject() string      { return "test-project" }
func (m *mockContextProvider) DefaultRegion() string       { return "test-region" }
func (m *mockContextProvider) DefaultOrganization() string { return "test-org" }

func setupClient(t *testing.T) (*Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"apiVersion":"` + builderAPIVersion + `","kind":"Bucket","metadata":{"id":"test"},"spec":{},"status":{}}`))
	}))

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	return NewClientWithParent(restClient, &mockContextProvider{}), server
}

func TestClient(t *testing.T) {
	ctx := context.Background()
	client, server := setupClient(t)
	defer server.Close()

	if client.Buckets() == nil || client.BucketServiceAccounts() == nil {
		t.Fatal("service getters failed")
	}

	req := NewBucketBuilder("test").Build()
	if _, err := client.Buckets().Create(ctx, req); err != nil {
		t.Errorf("Create: %v", err)
	}
	if _, err := client.Buckets().Get(ctx, "test"); err != nil {
		t.Errorf("Get: %v", err)
	}
	if _, err := client.Buckets().List(ctx); err != nil {
		t.Errorf("List: %v", err)
	}
	if err := client.Buckets().Delete(ctx, "test"); err != nil {
		t.Errorf("Delete: %v", err)
	}
	if _, err := client.Buckets().Patch(ctx, "test", map[string]interface{}{}); err != nil {
		t.Errorf("Patch: %v", err)
	}
	if _, err := client.Buckets().Update(ctx, "test", &storage.Bucket{ApiVersion: builderAPIVersion, Kind: "Bucket"}); err != nil {
		t.Errorf("Update: %v", err)
	}

	saReq := NewBucketServiceAccountBuilder("test").Build()
	if _, err := client.BucketServiceAccounts().Create(ctx, saReq); err != nil {
		t.Errorf("SA Create: %v", err)
	}
	if _, err := client.BucketServiceAccounts().Get(ctx, "test"); err != nil {
		t.Errorf("SA Get: %v", err)
	}
	if _, err := client.BucketServiceAccounts().List(ctx); err != nil {
		t.Errorf("SA List: %v", err)
	}
	if err := client.BucketServiceAccounts().Delete(ctx, "test"); err != nil {
		t.Errorf("SA Delete: %v", err)
	}
	if _, err := client.BucketServiceAccounts().Patch(ctx, "test", map[string]interface{}{}); err != nil {
		t.Errorf("SA Patch: %v", err)
	}
	if _, err := client.BucketServiceAccounts().Update(ctx, "test", &storage.BucketServiceAccount{ApiVersion: builderAPIVersion, Kind: "BucketServiceAccount"}); err != nil {
		t.Errorf("SA Update: %v", err)
	}

	if UpdateBucket("test", client.Buckets()) == nil {
		t.Error("UpdateBucket builder failed")
	}
	if UpdateBucketServiceAccount("test", client.BucketServiceAccounts()) == nil {
		t.Error("UpdateBucketServiceAccount builder failed")
	}
}
