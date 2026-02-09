// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package iam

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/types/iam"
)

type mockContextProvider struct{}

func (m *mockContextProvider) DefaultProject() string      { return "test-project" }
func (m *mockContextProvider) DefaultRegion() string       { return "test-region" }
func (m *mockContextProvider) DefaultOrganization() string { return "test-org" }

func setupClient(t *testing.T) (*Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		orgPath := "/iam/" + apiVersion + "/organizations"
		if r.URL.Path == orgPath+"/test" || r.URL.Path == orgPath {
			w.Write([]byte(`{"apiVersion":"` + builderAPIVersion + `","kind":"Organization","metadata":{"id":"550e8400-e29b-41d4-a716-446655440000"},"spec":{}}`))
		} else {
			w.Write([]byte(`{"apiVersion":"` + builderAPIVersion + `","kind":"Project","metadata":{"id":"test"},"spec":{},"status":{}}`))
		}
	}))

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	return NewClientWithParent(restClient, &mockContextProvider{}), server
}

func TestClient(t *testing.T) {
	ctx := context.Background()
	client, server := setupClient(t)
	defer server.Close()

	if client.Projects() == nil || client.PermissionSets() == nil || client.Organizations() == nil {
		t.Fatal("service getters failed")
	}

	projReq, _ := NewProjectBuilder("test", "org").Build()
	if _, err := client.Projects().Create(ctx, projReq); err != nil {
		t.Errorf("Project Create: %v", err)
	}
	if _, err := client.Projects().Get(ctx, "test"); err != nil {
		t.Errorf("Project Get: %v", err)
	}
	if _, err := client.Projects().List(ctx); err != nil {
		t.Errorf("Project List: %v", err)
	}
	if err := client.Projects().Delete(ctx, "test"); err != nil {
		t.Errorf("Project Delete: %v", err)
	}
	if _, err := client.Projects().Patch(ctx, "test", map[string]interface{}{}); err != nil {
		t.Errorf("Project Patch: %v", err)
	}
	if _, err := client.Projects().Update(ctx, "test", &iam.Project{ApiVersion: builderAPIVersion, Kind: "Project"}); err != nil {
		t.Errorf("Project Update: %v", err)
	}

	psReq := NewPermissionSetBuilder("test", "proj", "test@example.com").Build()
	if _, err := client.PermissionSets().Create(ctx, psReq); err != nil {
		t.Errorf("PS Create: %v", err)
	}
	if _, err := client.PermissionSets().Get(ctx, "test"); err != nil {
		t.Errorf("PS Get: %v", err)
	}
	if _, err := client.PermissionSets().List(ctx); err != nil {
		t.Errorf("PS List: %v", err)
	}
	if err := client.PermissionSets().Delete(ctx, "test"); err != nil {
		t.Errorf("PS Delete: %v", err)
	}

	if _, err := client.Organizations().Get(ctx, "test"); err != nil {
		t.Errorf("Org Get: %v", err)
	}
	if _, err := client.Organizations().List(ctx); err != nil {
		t.Errorf("Org List: %v", err)
	}
	if _, err := client.Organizations().Patch(ctx, "test", map[string]interface{}{}); err != nil {
		t.Errorf("Org Patch: %v", err)
	}
}
