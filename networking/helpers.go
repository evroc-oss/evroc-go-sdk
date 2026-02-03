// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package networking

import (
	networkingtypes "github.com/evroc-oss/evroc-go-sdk/types/networking"
)

// ============================================================================
// Status Helpers
// ============================================================================

// IsPublicIPReady returns true if the PublicIP is in Ready condition.
func IsPublicIPReady(ip *networkingtypes.PublicIP) bool {
	if ip == nil || ip.Status.Conditions == nil {
		return false
	}

	for _, cond := range *ip.Status.Conditions {
		if cond.Type == "Ready" && string(cond.Status) == "True" {
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
