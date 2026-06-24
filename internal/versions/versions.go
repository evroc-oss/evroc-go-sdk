// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package versions provides a central registry of API versions supported by this SDK.
// This is the single source of truth for all API version information.
package versions

import "fmt"

// SDKVersion is the semantic version of the SDK itself.
// Update this when releasing new SDK versions.
const SDKVersion = "0.6.0"

// APIVersion represents an API version string (e.g., "v1alpha2", "v1beta", "v1").
type APIVersion string

const (
	// Compute API version
	ComputeAPIVersion APIVersion = "v1beta2"

	// Networking API version
	NetworkingAPIVersion APIVersion = "v1beta2"

	// IAM API version
	IAMAPIVersion APIVersion = "v1beta1"

	// Storage API version
	StorageAPIVersion APIVersion = "v1"

	// Quotas API version
	QuotasAPIVersion APIVersion = "v1alpha2"

	// Think API version
	ThinkAPIVersion APIVersion = "v1beta2"

	// LoadBalancer API version (pre-release)
	LoadBalancerAPIVersion APIVersion = "v1alpha1"
)

// ServiceVersions maps service names to their API versions.
// This is used by the code generator and can be queried at runtime.
var ServiceVersions = map[string]APIVersion{
	"compute":    ComputeAPIVersion,
	"networking": NetworkingAPIVersion,
	"iam":        IAMAPIVersion,
	"storage":    StorageAPIVersion,
	"quotas":     QuotasAPIVersion,
	"think":        ThinkAPIVersion,
	"loadbalancer": LoadBalancerAPIVersion,
}

// APIVersionInfo contains detailed version information for all services.
type APIVersionInfo struct {
	SDKVersion   string
	Compute      string
	Networking   string
	IAM          string
	Storage      string
	Quotas       string
	Think        string
	LoadBalancer string
}

// Current returns the current API version information.
func Current() APIVersionInfo {
	return APIVersionInfo{
		SDKVersion:   SDKVersion,
		Compute:      string(ComputeAPIVersion),
		Networking:   string(NetworkingAPIVersion),
		IAM:          string(IAMAPIVersion),
		Storage:      string(StorageAPIVersion),
		Quotas:       string(QuotasAPIVersion),
		Think:        string(ThinkAPIVersion),
		LoadBalancer: string(LoadBalancerAPIVersion),
	}
}

// String returns a formatted string of all version information.
func (v APIVersionInfo) String() string {
	return fmt.Sprintf(
		"evroc Go SDK v%s\n"+
			"API Versions:\n"+
			"  compute:      %s\n"+
			"  networking:   %s\n"+
			"  iam:          %s\n"+
			"  storage:      %s\n"+
			"  quotas:       %s\n"+
			"  think:        %s\n"+
			"  loadbalancer: %s (pre-release)",
		v.SDKVersion,
		v.Compute,
		v.Networking,
		v.IAM,
		v.Storage,
		v.Quotas,
		v.Think,
		v.LoadBalancer,
	)
}

// GetAPIVersion returns the API version for a given service name.
// Returns empty string if service is not found.
func GetAPIVersion(service string) string {
	if v, ok := ServiceVersions[service]; ok {
		return string(v)
	}
	return ""
}
