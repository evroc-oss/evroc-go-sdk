// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"context"

	computetypes "github.com/evroc-oss/evroc-go-sdk/types/compute"
)

// CustomDiskImageBuilder registers an existing object as a custom image.
// It does not upload the object or create bucket credentials.
type CustomDiskImageBuilder struct {
	request computetypes.CustomDiskImageRequest
}

// NewCustomDiskImageBuilder creates an amd64 image registration using fully qualified
// bucket and bucket service account references. Set a default disk size before Build.
func NewCustomDiskImageBuilder(name, bucketRef, bucketServiceAccountRef, objectPath string) *CustomDiskImageBuilder {
	return &CustomDiskImageBuilder{request: computetypes.CustomDiskImageRequest{
		ApiVersion: builderAPIVersion, Kind: "CustomDiskImage",
		Metadata: computetypes.RegionalMetadataRequest{Id: name},
		Spec: computetypes.CustomDiskImageSpec{
			Architecture: computetypes.CustomDiskImageSpecArchitectureAmd64,
			Source:       computetypes.CustomDiskImageSpecSource{BucketRef: bucketRef, ServiceAccountRef: bucketServiceAccountRef, Path: objectPath},
		},
	}}
}

// WithArchitecture sets the image architecture supported by the API.
func (b *CustomDiskImageBuilder) WithArchitecture(architecture string) *CustomDiskImageBuilder {
	b.request.Spec.Architecture = computetypes.CustomDiskImageSpecArchitecture(architecture)
	return b
}

// WithDefaultDiskSize sets the required default disk capacity for disks using the image.
func (b *CustomDiskImageBuilder) WithDefaultDiskSize(amount int32, unit DiskSizeUnit) *CustomDiskImageBuilder {
	b.request.Spec.DefaultDiskSize = computetypes.CustomDiskImageSpecDefaultDiskSize{Amount: amount, Unit: computetypes.CustomDiskImageSpecDefaultDiskSizeUnit(unit)}
	return b
}

// WithDefaultDiskSizeGB sets the default disk capacity in gigabytes.
func (b *CustomDiskImageBuilder) WithDefaultDiskSizeGB(amount int32) *CustomDiskImageBuilder {
	return b.WithDefaultDiskSize(amount, DiskSizeUnitGB)
}

// WithObjectVersion pins the source object version. Without it, the API uses the latest object.
func (b *CustomDiskImageBuilder) WithObjectVersion(version string) *CustomDiskImageBuilder {
	b.request.Spec.Source.Version = &version
	return b
}

// WithDescription sets descriptive metadata; it does not alter the source image.
func (b *CustomDiskImageBuilder) WithDescription(description string) *CustomDiskImageBuilder {
	b.request.Spec.Description = &description
	return b
}

// WithOSName sets informational OS metadata.
func (b *CustomDiskImageBuilder) WithOSName(name string) *CustomDiskImageBuilder {
	b.request.Spec.OsName = &name
	return b
}

// WithOSVersion sets informational OS metadata.
func (b *CustomDiskImageBuilder) WithOSVersion(version string) *CustomDiskImageBuilder {
	b.request.Spec.OsVersion = &version
	return b
}

// WithImageVersion sets an informational image version; use WithObjectVersion to pin source bytes.
func (b *CustomDiskImageBuilder) WithImageVersion(version string) *CustomDiskImageBuilder {
	b.request.Spec.ImageVersion = &version
	return b
}

// WithLabels sets image labels.
func (b *CustomDiskImageBuilder) WithLabels(labels map[string]string) *CustomDiskImageBuilder {
	copied := computetypes.UserLabels{}
	for key, value := range labels {
		copied[key] = value
	}
	b.request.Metadata.UserLabels = &copied
	return b
}

// Build returns the image registration request. Required values are validated by the API.
func (b *CustomDiskImageBuilder) Build() *computetypes.CustomDiskImageRequest {
	result := b.request
	return &result
}

// Create builds and submits the image registration.
func (b *CustomDiskImageBuilder) Create(ctx context.Context, service *CustomDiskImagesService) (*computetypes.CustomDiskImage, error) {
	return service.Create(ctx, b.Build())
}
