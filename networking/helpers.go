// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"github.com/evroc-oss/evroc-go-sdk/metrics"
	networkingtypes "github.com/evroc-oss/evroc-go-sdk/types/networking"
)

// ============================================================================
// Metrics Support
// ============================================================================

// WithMetrics enables metrics collection for this networking client.
// Returns the client to allow chaining.
func (c *Client) WithMetrics(m *metrics.Manager) *Client {
	c.metrics = m
	return c
}

// ============================================================================
// Status Helpers
// ============================================================================

// IsPublicIPReady returns true if the PublicIP is in Ready condition.
func IsPublicIPReady(ip *networkingtypes.PublicIP) bool {
	if ip == nil || ip.Status.Conditions == nil {
		return false
	}

	for _, cond := range *ip.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}
	return false
}

// IsSecurityGroupReady returns true if the SecurityGroup is in Ready condition.
func IsSecurityGroupReady(sg *networkingtypes.SecurityGroup) bool {
	if sg == nil || sg.Status.Conditions == nil {
		return false
	}

	for _, cond := range *sg.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}
	return false
}

// GetPublicIPAddress extracts the actual IP address from a PublicIP resource.
func GetPublicIPAddress(ip *networkingtypes.PublicIP) string {
	if ip == nil || ip.Status.PublicIPv4Address == nil {
		return ""
	}
	return *ip.Status.PublicIPv4Address
}
