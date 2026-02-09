// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"apiVersion":"quotas/v1","kind":"Quota","items":[]}`))
	}))
	defer server.Close()

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})

	// Test NewClient (without parent)
	clientNoPar := NewClient(restClient)
	if clientNoPar == nil {
		t.Error("NewClient failed")
	}

	// Test NewClientWithParent
	client := NewClientWithParent(restClient, &mockContextProvider{})

	ctx := context.Background()
	if client.OrganizationQuotas() == nil || client.ProjectQuotas() == nil {
		t.Fatal("Quota service getters failed")
	}

	// Test Get with explicit IDs
	if _, err := client.ProjectQuotas().Get(ctx, "test-project"); err != nil {
		t.Errorf("Project Get failed: %v", err)
	}
	if _, err := client.OrganizationQuotas().Get(ctx, "test-org"); err != nil {
		t.Errorf("Org Get failed: %v", err)
	}

	// Generated services require explicit IDs (no defaulting behavior)
}
