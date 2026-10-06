// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute_test

import (
	"reflect"
	"testing"

	"github.com/evroc-oss/evroc-go-sdk/compute"
)

func TestTypedDiskImageCompatibility(t *testing.T) {
	for _, image := range []compute.DiskImage{compute.DiskImageUbuntu2404, "future-os", ""} {
		typed := compute.NewDiskBuilder("boot").WithDiskImage(image).Build()
		legacy := compute.NewDiskBuilder("boot").WithImage(string(image)).Build()
		if !reflect.DeepEqual(typed, legacy) {
			t.Errorf("image %q: typed and string builders differ", image)
		}
	}
}

func TestTypedComputeProfileCompatibility(t *testing.T) {
	for _, profile := range []compute.ComputeProfile{compute.VMSizeA1aM, "future-profile", ""} {
		typed := compute.NewVirtualMachineBuilder("vm").WithComputeProfile(profile).Build()
		legacy := compute.NewVirtualMachineBuilder("vm").WithVMInstanceType(string(profile)).Build()
		old := compute.NewVirtualMachineBuilder("vm").WithSize(string(profile)).Build()
		if !reflect.DeepEqual(typed, legacy) || !reflect.DeepEqual(typed, old) {
			t.Errorf("profile %q: typed and string builders differ", profile)
		}
	}
}
