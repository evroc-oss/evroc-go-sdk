// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
)

func TestServiceAccountCredentialRevoke(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	client := NewClient(restClient, &mockContextProvider{})

	if err := client.ServiceAccountCredentials("cicd").Revoke(context.Background(), "key-1", "leaked"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if want := "/iam/" + apiVersion + "/projects/test-project/serviceAccounts/cicd/serviceAccountCredentials/key-1/revoke"; gotPath != want {
		t.Errorf("expected path %q, got %q", want, gotPath)
	}
	if gotBody["reason"] != "leaked" {
		t.Errorf("expected reason %q, got %v", "leaked", gotBody["reason"])
	}
}

func TestServiceAccountCredentialRevokeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"reason":"NotFound"}`))
	}))
	defer server.Close()

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	client := NewClient(restClient, &mockContextProvider{})

	if err := client.ServiceAccountCredentials("cicd").Revoke(context.Background(), "missing", "leaked"); err == nil {
		t.Fatal("expected error for 404 response")
	}
}
