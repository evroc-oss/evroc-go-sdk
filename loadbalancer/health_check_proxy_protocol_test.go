// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package loadbalancer

import (
	"encoding/json"
	"testing"
)

func TestHealthCheckDisableProxyProtocol(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{"inherit", ""},
		{"disable", "true"},
		{"restore", "false"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := NewBackendServiceBuilder("svc").WithProxyProtocol(true)
			if tc.value != "" {
				b.WithHealthCheckDisableProxyProtocol(tc.value == "true")
			}
			// Selecting the check type after the override must preserve it.
			req := b.WithHTTPHealthCheck("/healthz").Build()
			if req.Spec.ProxyProtocol == nil || !*req.Spec.ProxyProtocol {
				t.Fatal("health-check override changed application PROXY protocol")
			}
			data, err := json.Marshal(req.Spec.HealthCheck)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(data, &fields); err != nil {
				t.Fatal(err)
			}
			if got := string(fields["disableProxyProtocol"]); got != tc.value {
				t.Fatalf("disableProxyProtocol = %q, want %q", got, tc.value)
			}
		})
	}
}
