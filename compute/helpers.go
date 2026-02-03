// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package compute

import (
	"errors"

	computetypes "github.com/evroc-oss/evroc-go-sdk/types/compute"
)

// ============================================================================
// Status Helpers
// ============================================================================

// IsAttachmentReady returns true if the HotswapDiskAttachment is in Ready condition.
func IsAttachmentReady(attachment *computetypes.HotswapDiskAttachment) bool {
	if attachment == nil || attachment.Status.Conditions == nil {
		return false
	}

	for _, cond := range *attachment.Status.Conditions {
		if cond.Type == "Ready" && string(cond.Status) == "True" {
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
		if cond.Type == "Ready" && string(cond.Status) == "True" {
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
		if cond.Type == "Ready" && string(cond.Status) == "True" {
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
			if string(cond.Status) == "True" {
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

// ValidateDiskSize validates that disk size is within acceptable ranges.
func ValidateDiskSize(amount int32, unit string) error {
	// Convert to GB for validation
	var sizeGB int32
	switch unit {
	case "KB":
		sizeGB = amount / (1024 * 1024)
	case "MB":
		sizeGB = amount / 1024
	case "GB":
		sizeGB = amount
	case "TB":
		sizeGB = amount * 1024
	default:
		return errors.New("invalid disk size unit, must be KB, MB, GB, or TB")
	}

	if sizeGB < 1 {
		return errors.New("disk size must be at least 1 GB")
	}

	if sizeGB > 10240 { // 10 TB max
		return errors.New("disk size cannot exceed 10 TB")
	}

	return nil
}

// ValidateVMSize validates that the VM size is a known valid size.
func ValidateVMSize(size string) error {
	// List of known VM sizes - this should be updated based on actual available sizes
	validSizes := map[string]bool{
		"a1a.xs": true,
		"a1a.s":  true,
		"a1a.m":  true,
		"a1a.l":  true,
		"a1a.xl": true,
		"c1a.xs": true,
		"c1a.s":  true,
		"c1a.m":  true,
		"c1a.l":  true,
		"c1a.xl": true,
		"m1a.xs": true,
		"m1a.s":  true,
		"m1a.m":  true,
		"m1a.l":  true,
		"m1a.xl": true,
	}

	if !validSizes[size] {
		return errors.New("invalid VM size, must be one of: a1a.xs, a1a.s, a1a.m, a1a.l, a1a.xl, c1a.xs, c1a.s, c1a.m, c1a.l, c1a.xl, m1a.xs, m1a.s, m1a.m, m1a.l, m1a.xl")
	}

	return nil
}


