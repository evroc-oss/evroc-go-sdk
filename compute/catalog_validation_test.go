// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
)

func TestCatalogValidation(t *testing.T) {
	for _, resource := range []string{"computeProfiles", "diskImages/evroc"} {
		for _, tc := range []struct {
			name, input, body string
			status            int
			wantErr           bool
		}{
			{"new API entry", "test-future-entry", `{"items":[{"metadata":{"id":"test-future-entry"}}]}`, 200, false},
			{"full reference", "/compute/global/" + resource + "/test-future-entry", `{"items":[{"metadata":{"id":"test-future-entry"}}]}`, 200, false},
			{"wrong reference collection", "/compute/global/other/test-future-entry", `{"items":[{"metadata":{"id":"test-future-entry"}}]}`, 200, true},
			{"unknown", "test-absent", `{"items":[{"metadata":{"id":"test-future-entry"}}]}`, 200, true},
			{"API failure", "test-future-entry", `{}`, 403, true},
		} {
			t.Run(resource+"/"+tc.name, func(t *testing.T) {
				calls := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++
					if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/global/"+resource) {
						t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tc.status)
					fmt.Fprint(w, tc.body)
				}))
				defer server.Close()
				rc, err := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
				if err != nil {
					t.Fatal(err)
				}
				cache := &CatalogCache{}
				for range 100 {
					c := NewClient(rc, &mockContextProvider{}).WithCatalogCache(cache)
					validate := c.ComputeProfiles().Validate
					if resource == "diskImages/evroc" {
						validate = c.DiskImages().Validate
					}
					err = validate(context.Background(), tc.input)
					if (err != nil) != tc.wantErr {
						t.Fatalf("Validate = %v, want error %v", err, tc.wantErr)
					}
					if tc.status == 403 && errors.Unwrap(err) == nil {
						t.Fatal("API error must be wrapped")
					}
				}
				wantCalls := 1
				if tc.status == 403 {
					wantCalls = 100
				} // Failed loads are retried.
				if calls != wantCalls {
					t.Fatalf("LIST calls = %d, want %d", calls, wantCalls)
				}

			})
		}
	}
}
