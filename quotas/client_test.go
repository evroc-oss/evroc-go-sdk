// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package quotas

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
)

type mockContextProvider struct{}

func (m *mockContextProvider) DefaultProject() string      { return "test-project" }
func (m *mockContextProvider) DefaultRegion() string       { return "test-region" }
func (m *mockContextProvider) DefaultOrganization() string { return "test-org" }

func TestClient(t *testing.T) {
	var lastPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"apiVersion":"quotas/v1","kind":"Quota","items":[]}`))
	}))
	defer server.Close()

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})

	client := NewClient(restClient, &mockContextProvider{})

	ctx := context.Background()
	if client.OrganizationQuotas() == nil || client.ProjectQuotas() == nil {
		t.Fatal("Quota service getters failed")
	}

	// Test ProjectQuotas.Get uses the correct regional path
	if _, err := client.ProjectQuotas().Get(ctx); err != nil {
		t.Errorf("Project Get failed: %v", err)
	}
	wantPath := "/quotas/v1alpha2/projects/test-project/regions/test-region/projectQuotas"
	if lastPath != wantPath {
		t.Errorf("ProjectQuotas.Get path = %q, want %q", lastPath, wantPath)
	}

	if _, err := client.OrganizationQuotas().Get(ctx); err != nil {
		t.Errorf("Org Get failed: %v", err)
	}
	wantOrgPath := "/quotas/v1alpha2/organizations/test-org/regions/test-region/orgQuotas"
	if lastPath != wantOrgPath {
		t.Errorf("OrganizationQuotas.Get path = %q, want %q", lastPath, wantOrgPath)
	}
}
