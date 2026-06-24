// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"github.com/evroc-oss/evroc-go-sdk/metrics"
	computetypes "github.com/evroc-oss/evroc-go-sdk/types/compute"
)

// ============================================================================
// Metrics Support
// ============================================================================

// WithMetrics enables metrics collection for this compute client.
// Returns the client to allow chaining.
func (c *Client) WithMetrics(m *metrics.Manager) *Client {
	c.metrics = m
	return c
}

// ============================================================================
// Status Helpers
// ============================================================================

// IsAttachmentReady returns true if the HotswapDiskAttachment is in Ready condition.
func IsAttachmentReady(attachment *computetypes.HotswapDiskAttachment) bool {
	if attachment == nil || attachment.Status.Conditions == nil {
		return false
	}

	for _, cond := range *attachment.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}

	return false
}

// IsVMReady returns true if the VirtualMachine is in Ready condition.
func IsVMReady(vm *computetypes.VirtualMachine) bool {
	if vm == nil || vm.Status.Conditions == nil {
		return false
	}

	for _, cond := range *vm.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}
	return false
}

// IsVMRunning returns true if the VirtualMachine spec indicates it should be running.
func IsVMRunning(vm *computetypes.VirtualMachine) bool {
	if vm == nil || vm.Spec.Running == nil {
		return true // Default is running
	}
	return *vm.Spec.Running
}

// IsDiskReady returns true if the Disk is in Ready condition.
func IsDiskReady(disk *computetypes.Disk) bool {
	if disk == nil || disk.Status.Conditions == nil {
		return false
	}

	for _, cond := range *disk.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}
	return false
}

// IsPlacementGroupReady returns true if the PlacementGroup is in Ready condition.
func IsPlacementGroupReady(pg *computetypes.PlacementGroup) bool {
	if pg == nil || pg.Status.Conditions == nil {
		return false
	}

	for _, cond := range *pg.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}
	return false
}

// GetVMState returns the current state of a VirtualMachine based on its conditions.
func GetVMState(vm *computetypes.VirtualMachine) string {
	if vm == nil || vm.Status.Conditions == nil {
		return "Unknown"
	}

	// Check if VM is running based on spec
	if vm.Spec.Running != nil && !*vm.Spec.Running {
		return "Stopped"
	}

	// Check conditions for Ready state
	for _, cond := range *vm.Status.Conditions {
		if cond.Type == "Ready" {
			if cond.Status == "True" {
				return "Running"
			}
			// Return reason if not ready
			if cond.Reason != "" {
				return cond.Reason
			}
		}
	}

	return "Pending"
}
