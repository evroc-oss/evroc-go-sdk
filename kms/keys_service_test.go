// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package kms

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
)

type mockContextProvider struct{}

func (m *mockContextProvider) DefaultProject() string      { return "test-project" }
func (m *mockContextProvider) DefaultRegion() string       { return "test-region" }
func (m *mockContextProvider) DefaultOrganization() string { return "test-org" }

func TestKeyActions(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotBody = r.Method, r.URL.Path, nil
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	keys := NewClient(restClient, &mockContextProvider{}).Keys()
	ctx := context.Background()

	tests := []struct {
		action   string
		call     func() error
		wantBody map[string]any
	}{
		{"activate", func() error { return keys.Activate(ctx, "k1") }, nil},
		{"suspend", func() error { return keys.Suspend(ctx, "k1") }, nil},
		{"abortDeletion", func() error { return keys.AbortDeletion(ctx, "k1") }, nil},
		{"trim", func() error { return keys.Trim(ctx, "k1", 3) }, map[string]any{"minimumVersionToKeep": float64(3)}},
		{"abortTrimming", func() error { return keys.AbortTrimming(ctx, "k1") }, nil},
		{"rotate", func() error { return keys.Rotate(ctx, "k1") }, nil},
		{"import", func() error {
			return keys.Import(ctx, "k1", &ImportRequest{Hash: HashSHA256, WrappedDataKey: []byte("w"), MaterialCiphertext: []byte("m")})
		}, map[string]any{"hash": "SHA256", "wrappedDataKey": "dw==", "materialCiphertext": "bQ=="}},
		{"rotateImported", func() error {
			return keys.RotateImported(ctx, "k1", &RotateImportedRequest{Hash: HashSHA256, WrappedDataKey: []byte("w"), MaterialCiphertext: []byte("m")})
		}, map[string]any{"hash": "SHA256", "wrappedDataKey": "dw==", "materialCiphertext": "bQ=="}},
	}
	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			if err := tt.call(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotMethod != http.MethodPost {
				t.Errorf("expected POST, got %s", gotMethod)
			}
			if want := "/kms/v1/projects/test-project/regions/test-region/keys/k1/" + tt.action; gotPath != want {
				t.Errorf("expected path %q, got %q", want, gotPath)
			}
			for k, v := range tt.wantBody {
				if gotBody[k] != v {
					t.Errorf("body[%q] = %v, want %v", k, gotBody[k], v)
				}
			}
		})
	}
}

func TestKeyActionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"reason":"NotFound"}`))
	}))
	defer server.Close()

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	if err := NewClient(restClient, &mockContextProvider{}).Keys().Rotate(context.Background(), "missing"); err == nil {
		t.Fatal("expected error for 404 response")
	}
}
