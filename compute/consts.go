// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	computetypes "github.com/evroc-oss/evroc-go-sdk/types/compute"
)

// Re-export commonly used constants from types/compute for easier access.
// This allows users to write compute.Hotswap instead of computetypes.Hotswap.

// Disk Usage Kind constants
const (
	// Hotswap indicates the disk is attached via hotswap and can be detached while VM is running.
	Hotswap = string(computetypes.Hotswap)

	// Permanent indicates the disk is permanently attached to the VM.
	Permanent = string(computetypes.Permanent)
)

// Placement Group Strategy constants
const (
	// Spread ensures VMs in the placement group are spread across different physical hosts
	// for high availability.
	Spread = string(computetypes.Spread)
)

// Disk Size Unit constants
const (
	KB = string(computetypes.DiskSpecDiskSizeUnitKB)
	MB = string(computetypes.DiskSpecDiskSizeUnitMB)
	GB = string(computetypes.DiskSpecDiskSizeUnitGB)
	TB = string(computetypes.DiskSpecDiskSizeUnitTB)
)
