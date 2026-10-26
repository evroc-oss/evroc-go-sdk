// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
)

func setupConsoleServer(t *testing.T, response string, gotMethod, gotPath *string) (*Client, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*gotMethod, *gotPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	return NewClient(restClient, &mockContextProvider{}), server.Close
}

func TestVirtualMachineConsoleLogs(t *testing.T) {
	var gotMethod, gotPath string
	client, closeFn := setupConsoleServer(t, `{"logs":["line 1","line 2"]}`, &gotMethod, &gotPath)
	defer closeFn()

	logs, err := client.VirtualMachines().ConsoleLogs(context.Background(), "vm-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	if want := "/compute/" + apiVersion + "/projects/test-project/regions/test-region/virtualMachines/vm-1/logs/console"; gotPath != want {
		t.Errorf("expected path %q, got %q", want, gotPath)
	}
	if len(logs.Logs) != 2 || logs.Logs[1] != "line 2" {
		t.Errorf("unexpected logs: %v", logs.Logs)
	}
}
