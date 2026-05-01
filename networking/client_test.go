// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/types/networking"
)

type mockContextProvider struct{}

func (m *mockContextProvider) DefaultProject() string      { return "test-project" }
func (m *mockContextProvider) DefaultRegion() string       { return "test-region" }
func (m *mockContextProvider) DefaultOrganization() string { return "test-org" }

func setupClient(t *testing.T) (*Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"apiVersion":"` + builderAPIVersion + `","kind":"PublicIP","metadata":{"id":"test"},"spec":{},"status":{}}`))
	}))

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	return NewClient(restClient, &mockContextProvider{}), server
}

func TestClient(t *testing.T) {
	ctx := context.Background()
	client, server := setupClient(t)
	defer server.Close()

	if client.PublicIPs() == nil || client.SecurityGroups() == nil {
		t.Fatal("service getters failed")
	}

	req := NewPublicIPBuilder("test").Build()
	if _, err := client.PublicIPs().Create(ctx, req); err != nil {
		t.Errorf("Create: %v", err)
	}
	if _, err := client.PublicIPs().Get(ctx, "test"); err != nil {
		t.Errorf("Get: %v", err)
	}
	if _, err := client.PublicIPs().List(ctx); err != nil {
		t.Errorf("List: %v", err)
	}
	if err := client.PublicIPs().Delete(ctx, "test"); err != nil {
		t.Errorf("Delete: %v", err)
	}
	if _, err := client.PublicIPs().Patch(ctx, "test", map[string]interface{}{}); err != nil {
		t.Errorf("Patch: %v", err)
	}
	if _, err := client.PublicIPs().Patch(ctx, "test", &networking.PublicIP{ApiVersion: builderAPIVersion, Kind: "PublicIP"}); err != nil {
		t.Errorf("Update: %v", err)
	}

	sgReq := NewSecurityGroupBuilder("test").Build()
	if _, err := client.SecurityGroups().Create(ctx, sgReq); err != nil {
		t.Errorf("SG Create: %v", err)
	}
	if _, err := client.SecurityGroups().Get(ctx, "test"); err != nil {
		t.Errorf("SG Get: %v", err)
	}
	if _, err := client.SecurityGroups().List(ctx); err != nil {
		t.Errorf("SG List: %v", err)
	}

	if client.Subnets() == nil || client.VirtualPrivateClouds() == nil {
		t.Fatal("additional service getters failed")
	}
}
