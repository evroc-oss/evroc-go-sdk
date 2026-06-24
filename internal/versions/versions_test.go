// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package versions

import (
	"strings"
	"testing"
)

func TestGetAPIVersion(t *testing.T) {
	tests := []struct {
		name     string
		service  string
		expected string
	}{
		{"compute", "compute", "v1beta2"},
		{"networking", "networking", "v1beta2"},
		{"iam", "iam", "v1beta1"},
		{"storage", "storage", "v1"},
		{"quotas", "quotas", "v1alpha2"},
		{"unknown service", "unknown", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAPIVersion(tt.service)
			if result != tt.expected {
				t.Errorf("GetAPIVersion(%q) = %q, expected %q", tt.service, result, tt.expected)
			}
		})
	}
}

func TestCurrent(t *testing.T) {
	info := Current()

	if info.SDKVersion != SDKVersion {
		t.Errorf("Current().SDKVersion = %q, expected %q", info.SDKVersion, SDKVersion)
	}

	if info.Compute != string(ComputeAPIVersion) {
		t.Errorf("Current().Compute = %q, expected %q", info.Compute, string(ComputeAPIVersion))
	}

	if info.Networking != string(NetworkingAPIVersion) {
		t.Errorf("Current().Networking = %q, expected %q", info.Networking, string(NetworkingAPIVersion))
	}

	if info.IAM != string(IAMAPIVersion) {
		t.Errorf("Current().IAM = %q, expected %q", info.IAM, string(IAMAPIVersion))
	}

	if info.Storage != string(StorageAPIVersion) {
		t.Errorf("Current().Storage = %q, expected %q", info.Storage, string(StorageAPIVersion))
	}

	if info.Quotas != string(QuotasAPIVersion) {
		t.Errorf("Current().Quotas = %q, expected %q", info.Quotas, string(QuotasAPIVersion))
	}
}

func TestAPIVersionInfoString(t *testing.T) {
	info := Current()
	result := info.String()

	// Check that the string contains expected elements
	expectedParts := []string{
		"evroc Go SDK",
		SDKVersion,
		"API Versions:",
		"compute:",
		"networking:",
		"iam:",
		"storage:",
		"quotas:",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("String() output missing expected part %q\nGot: %s", part, result)
		}
	}
}

func TestServiceVersionsMap(t *testing.T) {
	// Verify all expected services are in the map
	expectedServices := []string{"compute", "networking", "iam", "storage", "quotas", "think", "loadbalancer"}

	for _, service := range expectedServices {
		if _, ok := ServiceVersions[service]; !ok {
			t.Errorf("ServiceVersions missing expected service %q", service)
		}
	}

	// Verify map has exactly the expected number of services
	if len(ServiceVersions) != len(expectedServices) {
		t.Errorf("ServiceVersions has %d entries, expected %d",
			len(ServiceVersions), len(expectedServices))
	}
}

func TestSDKVersionFormat(t *testing.T) {
	// Verify SDK version follows semver format (basic check)
	if SDKVersion == "" {
		t.Error("SDKVersion should not be empty")
	}

	// Should contain at least one dot (x.y format minimum)
	if !strings.Contains(SDKVersion, ".") {
		t.Errorf("SDKVersion %q does not appear to be in semver format", SDKVersion)
	}
}

func TestAPIVersionConstants(t *testing.T) {
	// Verify all API version constants are non-empty
	tests := []struct {
		name    string
		version APIVersion
	}{
		{"ComputeAPIVersion", ComputeAPIVersion},
		{"NetworkingAPIVersion", NetworkingAPIVersion},
		{"IAMAPIVersion", IAMAPIVersion},
		{"StorageAPIVersion", StorageAPIVersion},
		{"QuotasAPIVersion", QuotasAPIVersion},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.version == "" {
				t.Errorf("%s should not be empty", tt.name)
			}
			// Verify it starts with 'v' (common API version convention)
			if !strings.HasPrefix(string(tt.version), "v") {
				t.Errorf("%s = %q, should start with 'v'", tt.name, tt.version)
			}
		})
	}
}
