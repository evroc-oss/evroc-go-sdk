// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import "testing"

func TestCustomDiskImageRef(t *testing.T) {
	project, region, empty := "project", "se-sto", ""
	for _, tc := range []struct {
		name  string
		image *CustomDiskImage
		want  string
	}{
		{"nil", nil, ""},
		{"missing metadata", &CustomDiskImage{}, ""},
		{"missing project", &CustomDiskImage{Metadata: RegionalMetadataResponse{Id: "image", Region: &region}}, ""},
		{"empty project", &CustomDiskImage{Metadata: RegionalMetadataResponse{Id: "image", Project: &empty, Region: &region}}, ""},
		{"missing region", &CustomDiskImage{Metadata: RegionalMetadataResponse{Id: "image", Project: &project}}, ""},
		{"valid", &CustomDiskImage{Metadata: RegionalMetadataResponse{Id: "image", Project: &project, Region: &region}}, "/compute/projects/project/regions/se-sto/customDiskImages/image"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.image.Ref().String(); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
