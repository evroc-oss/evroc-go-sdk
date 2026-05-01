// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import "testing"

func TestFindVMInstanceTypeEdgeCases(t *testing.T) {
	// Test various CPU/memory combinations
	cases := []struct {
		cpu    int
		memory int
		gpu    int
		want   string
	}{
		{2, 4, 0, "c1a.s"},
		{4, 8, 0, "c1a.m"},
		{8, 16, 0, "c1a.l"},
		{16, 32, 0, "c1a.xl"},
		{1, 1, 1, ""},     // GPU request
		{999, 999, 0, ""}, // No match
	}

	for _, tc := range cases {
		got := FindVMInstanceType(tc.cpu, tc.memory, tc.gpu)
		if tc.want == "" && got != "" {
			t.Errorf("FindVMInstanceType(%d, %d, %d) = %q, want empty", tc.cpu, tc.memory, tc.gpu, got)
		} else if tc.want != "" && got == "" {
			t.Errorf("FindVMInstanceType(%d, %d, %d) = empty, want %q", tc.cpu, tc.memory, tc.gpu, tc.want)
		}
	}
}

func TestAllDiskImages(t *testing.T) {
	images := []DiskImage{DiskImageUbuntuMinimal2404, DiskImageUbuntu2404, DiskImageUbuntu2204, DiskImageRocky100, DiskImageRocky96, DiskImageRocky95, DiskImageRocky810, DiskImageOpenSUSE156, DiskImageOpenSUSE155, DiskImageSLES156, DiskImageSLES155, DiskImageSLMicro61}

	for _, img := range images {
		if !IsValidDiskImage(string(img)) {
			t.Errorf("IsValidDiskImage(%s) = false, want true", img)
		}
	}
	if IsValidDiskImage("invalid") {
		t.Error("IsValidDiskImage should reject invalid")
	}
	if output := GetValidDiskImagesString(); len(output) < 50 {
		t.Error("GetValidDiskImagesString returned unexpectedly short string")
	}
}

func TestAllVMSizes(t *testing.T) {
	sizes := []string{"c1a.s", "c1a.m", "c1a.l", "c1a.xl", "c1a.2xl", "m1a.s", "m1a.m", "m1a.l", "m1a.xl"}

	for _, size := range sizes {
		if !IsValidVMSize(size) {
			t.Errorf("IsValidVMSize(%s) = false, want true", size)
		}
	}
	if IsValidVMSize("invalid") {
		t.Error("IsValidVMSize should reject invalid")
	}
	if output := GetValidVMSizesString(); len(output) < 50 {
		t.Error("GetValidVMSizesString returned unexpectedly short string")
	}
}
