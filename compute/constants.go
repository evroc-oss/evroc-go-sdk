// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package compute constants contains all valid disk images and VM compute profiles
// supported by the evroc platform. Update this file when new images or profiles
// become available.
package compute

import "strings"

// ============================================================================
// Disk Size Units
// ============================================================================

// DiskSizeUnit represents a disk size unit.
type DiskSizeUnit string

const (
	// DiskSizeUnitKB represents kilobytes
	DiskSizeUnitKB DiskSizeUnit = "KB"
	// DiskSizeUnitGB represents gigabytes
	DiskSizeUnitGB DiskSizeUnit = "GB"
)

// ============================================================================
// Disk Images
// ============================================================================

// DiskImage represents a valid OS image name.
type DiskImage string

const (
	// Ubuntu images
	DiskImageUbuntuMinimal2404 DiskImage = "ubuntu-minimal.24-04.1"
	DiskImageUbuntu2404        DiskImage = "ubuntu.24-04.1"
	DiskImageUbuntu2204        DiskImage = "ubuntu.22-04.1"

	// Rocky Linux images
	DiskImageRocky100 DiskImage = "rocky.10-0.1"
	DiskImageRocky96  DiskImage = "rocky.9-6.1"
	DiskImageRocky95  DiskImage = "rocky.9-5.1"
	DiskImageRocky810 DiskImage = "rocky.8-10.1"

	// openSUSE images
	DiskImageOpenSUSE156 DiskImage = "opensuse.15-6.1"
	DiskImageOpenSUSE155 DiskImage = "opensuse.15-5.1"

	// SUSE Linux Enterprise Server images
	DiskImageSLES156 DiskImage = "sles.15-6.1"
	DiskImageSLES155 DiskImage = "sles.15-5.1"

	// SL Micro images
	DiskImageSLMicro61 DiskImage = "sl-micro.6-1.1"
)

// ValidDiskImages contains all supported disk images.
var ValidDiskImages = []DiskImage{
	DiskImageUbuntuMinimal2404,
	DiskImageUbuntu2404,
	DiskImageUbuntu2204,
	DiskImageRocky100,
	DiskImageRocky96,
	DiskImageRocky95,
	DiskImageRocky810,
	DiskImageOpenSUSE156,
	DiskImageOpenSUSE155,
	DiskImageSLES156,
	DiskImageSLES155,
	DiskImageSLMicro61,
}

// validDiskImagesMap provides O(1) lookup for image validation.
var validDiskImagesMap map[string]bool

func init() {
	validDiskImagesMap = make(map[string]bool, len(ValidDiskImages))
	for _, img := range ValidDiskImages {
		validDiskImagesMap[string(img)] = true
	}
}

// IsValidDiskImage returns true if the given image name is valid.
func IsValidDiskImage(image string) bool {
	return validDiskImagesMap[image]
}

// GetValidDiskImagesString returns a comma-separated string of all valid disk images.
func GetValidDiskImagesString() string {
	images := make([]string, len(ValidDiskImages))
	for i, img := range ValidDiskImages {
		images[i] = string(img)
	}
	return strings.Join(images, ", ")
}

// ============================================================================
// VM Compute Profiles (VM Sizes)
// ============================================================================

// VMSize represents a valid VM compute profile.
type VMSize string

// A-series: General-purpose VMs (1:4 CPU:Memory ratio)
const (
	VMSizeA1aXS  VMSize = "a1a.xs"
	VMSizeA1aS   VMSize = "a1a.s"
	VMSizeA1aM   VMSize = "a1a.m"
	VMSizeA1aL   VMSize = "a1a.l"
	VMSizeA1aXL  VMSize = "a1a.xl"
	VMSizeA1a2XL VMSize = "a1a.2xl"
)

// C-series: Compute-optimized VMs (1:2 CPU:Memory ratio)
const (
	VMSizeC1aS   VMSize = "c1a.s"
	VMSizeC1aM   VMSize = "c1a.m"
	VMSizeC1aL   VMSize = "c1a.l"
	VMSizeC1aXL  VMSize = "c1a.xl"
	VMSizeC1a2XL VMSize = "c1a.2xl"
)

// M-series: Memory-optimized VMs (1:8 CPU:Memory ratio)
const (
	VMSizeM1aS  VMSize = "m1a.s"
	VMSizeM1aM  VMSize = "m1a.m"
	VMSizeM1aL  VMSize = "m1a.l"
	VMSizeM1aXL VMSize = "m1a.xl"
)

// GPU-enabled VMs: NVIDIA L40S series
const (
	VMSizeGnL40sS VMSize = "gn-l40s.s" // 15 vCPUs, 198 GB, 1 GPU, 3.8 TB NVMe
	VMSizeGnL40sM VMSize = "gn-l40s.m" // 30 vCPUs, 396 GB, 2 GPUs, 7.6 TB NVMe
	VMSizeGnL40sL VMSize = "gn-l40s.l" // 60 vCPUs, 792 GB, 4 GPUs, 15.2 TB NVMe
)

// GPU-enabled VMs: NVIDIA B200 series
const (
	VMSizeGnB200S  VMSize = "gn-b200.s"  // 26 vCPUs, 262 GB, 1 GPU, 4 TB NVMe
	VMSizeGnB200M  VMSize = "gn-b200.m"  // 52 vCPUs, 524 GB, 2 GPUs, 8 TB NVMe
	VMSizeGnB200L  VMSize = "gn-b200.l"  // 104 vCPUs, 1048 GB, 4 GPUs, 16 TB NVMe
	VMSizeGnB200XL VMSize = "gn-b200.xl" // 208 vCPUs, 2096 GB, 8 GPUs, 32 TB NVMe
)

// VMSizeSeries represents a family of VM sizes.
type VMSizeSeries struct {
	Name        string
	Description string
	Sizes       []VMSize
}

// VMSizeSeriesA1a contains all A-series (general-purpose) VM sizes.
var VMSizeSeriesA1a = VMSizeSeries{
	Name:        "a1a",
	Description: "General-purpose VMs (1:4 CPU:Memory)",
	Sizes: []VMSize{
		VMSizeA1aXS,
		VMSizeA1aS,
		VMSizeA1aM,
		VMSizeA1aL,
		VMSizeA1aXL,
		VMSizeA1a2XL,
	},
}

// VMSizeSeriesC1a contains all C-series (compute-optimized) VM sizes.
var VMSizeSeriesC1a = VMSizeSeries{
	Name:        "c1a",
	Description: "Compute-optimized VMs (1:2 CPU:Memory)",
	Sizes: []VMSize{
		VMSizeC1aS,
		VMSizeC1aM,
		VMSizeC1aL,
		VMSizeC1aXL,
		VMSizeC1a2XL,
	},
}

// VMSizeSeriesM1a contains all M-series (memory-optimized) VM sizes.
var VMSizeSeriesM1a = VMSizeSeries{
	Name:        "m1a",
	Description: "Memory-optimized VMs (1:8 CPU:Memory)",
	Sizes: []VMSize{
		VMSizeM1aS,
		VMSizeM1aM,
		VMSizeM1aL,
		VMSizeM1aXL,
	},
}

// VMSizeSeriesGnL40s contains all NVIDIA L40S GPU VM sizes.
var VMSizeSeriesGnL40s = VMSizeSeries{
	Name:        "gn-l40s",
	Description: "NVIDIA L40S GPU VMs with NVMe storage",
	Sizes: []VMSize{
		VMSizeGnL40sS,
		VMSizeGnL40sM,
		VMSizeGnL40sL,
	},
}

// VMSizeSeriesGnB200 contains all NVIDIA B200 GPU VM sizes.
var VMSizeSeriesGnB200 = VMSizeSeries{
	Name:        "gn-b200",
	Description: "NVIDIA B200 GPU VMs with NVMe storage",
	Sizes: []VMSize{
		VMSizeGnB200S,
		VMSizeGnB200M,
		VMSizeGnB200L,
		VMSizeGnB200XL,
	},
}

// AllVMSizeSeries contains all VM size series.
var AllVMSizeSeries = []VMSizeSeries{
	VMSizeSeriesA1a,
	VMSizeSeriesC1a,
	VMSizeSeriesM1a,
	VMSizeSeriesGnL40s,
	VMSizeSeriesGnB200,
}

// ValidVMSizes contains all supported VM sizes.
var ValidVMSizes []VMSize

// validVMSizesMap provides O(1) lookup for VM size validation.
var validVMSizesMap map[string]bool

func init() {
	// Build the flat list of all valid VM sizes
	for _, series := range AllVMSizeSeries {
		ValidVMSizes = append(ValidVMSizes, series.Sizes...)
	}

	// Build the lookup map
	validVMSizesMap = make(map[string]bool, len(ValidVMSizes))
	for _, size := range ValidVMSizes {
		validVMSizesMap[string(size)] = true
	}
}

// IsValidVMSize returns true if the given VM size is valid.
func IsValidVMSize(size string) bool {
	return validVMSizesMap[size]
}

// GetValidVMSizesString returns a comma-separated string of all valid VM sizes.
func GetValidVMSizesString() string {
	sizes := make([]string, len(ValidVMSizes))
	for i, size := range ValidVMSizes {
		sizes[i] = string(size)
	}
	return strings.Join(sizes, ", ")
}

// ============================================================================
// VM Size Specifications (Formula-based)
// ============================================================================

var (
	// CPU:Memory ratios for each series
	ratioToSeries = map[int]string{
		4: "a1a", // General-purpose
		2: "c1a", // Compute-optimized
		8: "m1a", // Memory-optimized
	}

	// vCPUs to size tier mapping
	vcpusToTier = map[int]string{
		1:  "xs",
		2:  "s",
		4:  "m",
		8:  "l",
		16: "xl",
		32: "2xl",
	}

	// GPU specs: map[gpuCount]map[vcpus]size
	gpuL40sSpecs = map[int]map[int]string{
		1: {15: "gn-l40s.s"},
		2: {30: "gn-l40s.m"},
		4: {60: "gn-l40s.l"},
	}

	gpuB200Specs = map[int]map[int]string{
		1: {26: "gn-b200.s"},
		2: {52: "gn-b200.m"},
		4: {104: "gn-b200.l"},
		8: {208: "gn-b200.xl"},
	}
)

// FindVMInstanceType finds a VM compute profile that exactly matches the CPU and memory requirements.
// For CPU-only VMs, uses the CPU:Memory ratios: a1a (1:4), c1a (1:2), m1a (1:8).
// Returns empty string if no exact match exists.
//
// Note: Function name says "InstanceType" for backwards compatibility, but returns a compute profile.
//
// Example:
//
//	FindVMInstanceType(4, 16, 0)  // Returns "a1a.m"
//	FindVMInstanceType(4, 8, 0)   // Returns "c1a.m"
//	FindVMInstanceType(2, 16, 0)  // Returns "m1a.s"
func FindVMInstanceType(vcpus, memoryGB, gpus int) string {
	if memoryGB == 0 || vcpus == 0 {
		return ""
	}

	// Handle GPU VMs
	if gpus > 0 {
		return findGPUVMSize(vcpus, gpus)
	}

	// Handle CPU-only VMs
	ratio := memoryGB / vcpus
	series, ok := ratioToSeries[ratio]
	if !ok {
		return ""
	}

	tier, ok := vcpusToTier[vcpus]
	if !ok {
		return ""
	}

	size := series + "." + tier
	if IsValidVMSize(size) {
		return size
	}
	return ""
}

// findGPUVMSize finds a GPU VM size by vCPUs and GPU count.
func findGPUVMSize(vcpus, gpus int) string {
	// Try L40S first
	if cpuMap, ok := gpuL40sSpecs[gpus]; ok {
		if size, ok := cpuMap[vcpus]; ok {
			return size
		}
	}

	// Try B200
	if cpuMap, ok := gpuB200Specs[gpus]; ok {
		if size, ok := cpuMap[vcpus]; ok {
			return size
		}
	}

	return ""
}
