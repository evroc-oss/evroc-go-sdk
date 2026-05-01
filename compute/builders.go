// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package compute provides builder patterns for compute resources.
package compute

import (
	"context"
	"log"

	compute "github.com/evroc-oss/evroc-go-sdk/types/compute"
)

// builderAPIVersion is the full API version string for compute resource requests.
// It references the apiVersion constant from client_generated.go.
const builderAPIVersion = "compute/" + apiVersion

// Default values for compute resource builders
const (
	// DefaultDiskSizeGB is the default disk size in GB when not specified
	DefaultDiskSizeGB = 10
	// DefaultVMInstanceType is the default VM compute profile when not specified
	DefaultVMInstanceType = "a1a.xs"
)

// DiskBuilder provides a fluent interface for creating Disk resources.
type DiskBuilder struct {
	id         string
	image      string
	sizeAmount int32
	sizeUnit   string
	zone       string
	labels     map[string]string
}

// NewDiskBuilder creates a new DiskBuilder with sensible defaults.
func NewDiskBuilder(id string) *DiskBuilder {
	return &DiskBuilder{
		id:         id,
		sizeAmount: DefaultDiskSizeGB,
		sizeUnit:   "GB",
	}
}

// WithImage sets the disk image (e.g., "ubuntu-minimal.24-04.1", "ubuntu.24-04.1", "rocky.10-0.1").
func (b *DiskBuilder) WithImage(image string) *DiskBuilder {
	b.image = image
	return b
}

// WithSize sets the disk size with a specific unit.
func (b *DiskBuilder) WithSize(amount int32, unit DiskSizeUnit) *DiskBuilder {
	b.sizeAmount = amount
	b.sizeUnit = string(unit)
	return b
}

// WithSizeGB is a convenience method to set size in GB.
func (b *DiskBuilder) WithSizeGB(sizeGB int32) *DiskBuilder {
	return b.WithSize(sizeGB, DiskSizeUnitGB)
}

// WithZone sets the availability zone for the disk.
func (b *DiskBuilder) WithZone(zone string) *DiskBuilder {
	b.zone = zone
	return b
}

// WithLabels sets user-defined labels for the disk.
func (b *DiskBuilder) WithLabels(labels map[string]string) *DiskBuilder {
	b.labels = labels
	return b
}

// Build creates the DiskRequest structure ready for the Create API call.
func (b *DiskBuilder) Build() *compute.DiskRequest {
	diskReq := &compute.DiskRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "Disk",
		Metadata: compute.RegionalMetadataRequest{
			Id: b.id,
		},
		Spec: compute.DiskSpec{
			Placement: compute.DiskSpecPlacement{}, // Always required, non-pointer
		},
	}

	// Add optional image
	if b.image != "" {
		imageRef := "/compute/global/diskImages/evroc/" + b.image
		diskReq.Spec.DiskImageRef = &imageRef
	}

	// Add custom disk size if specified
	if b.sizeAmount > 0 {
		diskReq.Spec.DiskSize = &compute.DiskSpecDiskSize{
			Amount: b.sizeAmount,
			Unit:   compute.DiskSpecDiskSizeUnit(b.sizeUnit),
		}
	}

	// Set placement zone (mandatory)
	diskReq.Spec.Placement.Zone = &b.zone

	// Add labels if specified
	if len(b.labels) > 0 {
		userLabels := compute.UserLabels(b.labels)
		diskReq.Metadata.UserLabels = &userLabels
	}

	return diskReq
}

// Create is a convenience method that builds and creates the disk in one call.
func (b *DiskBuilder) Create(ctx context.Context, client *DisksService) (*compute.Disk, error) {
	diskReq := b.Build()
	return client.Create(ctx, diskReq)
}

// VirtualMachineBuilder provides a fluent interface for creating VirtualMachine resources.
type VirtualMachineBuilder struct {
	id             string
	diskRefs       []diskRef
	vmSize         string
	publicIP       PublicIPRef
	securityGroups []SecurityGroupRef
	sshKeys        []string
	cloudInitData  string
	zone           string
	placementGroup PlacementGroupRef
	running        *bool
	labels         map[string]string
}

type diskRef struct {
	ref      DiskRef
	bootFrom bool
}

// NewVirtualMachineBuilder creates a new VirtualMachineBuilder with sensible defaults.
func NewVirtualMachineBuilder(id string) *VirtualMachineBuilder {
	running := true
	return &VirtualMachineBuilder{
		id:       id,
		vmSize:   DefaultVMInstanceType,
		diskRefs: []diskRef{},
		running:  &running,
	}
}

// WithBootDisk accepts a type-safe DiskRef (FQID only).
// Use disk.Ref() for resource chaining.
func (b *VirtualMachineBuilder) WithBootDisk(ref DiskRef) *VirtualMachineBuilder {
	b.diskRefs = append(b.diskRefs, diskRef{
		ref:      ref,
		bootFrom: true,
	})
	return b
}

// WithDataDisk accepts a type-safe DiskRef (FQID only).
// Use disk.Ref() for resource chaining.
func (b *VirtualMachineBuilder) WithDataDisk(ref DiskRef) *VirtualMachineBuilder {
	b.diskRefs = append(b.diskRefs, diskRef{
		ref:      ref,
		bootFrom: false,
	})
	return b
}

// WithVMInstanceType sets the VM compute profile (e.g., "a1a.xs", "c1a.m", "m1a.l").
// Note: Function name says "InstanceType" for backwards compatibility, but this sets the compute profile.
func (b *VirtualMachineBuilder) WithVMInstanceType(profile string) *VirtualMachineBuilder {
	b.vmSize = profile
	return b
}

// WithSize is deprecated. Use WithVMInstanceType instead.
func (b *VirtualMachineBuilder) WithSize(size string) *VirtualMachineBuilder {
	return b.WithVMInstanceType(size)
}

// WithPublicIP accepts a type-safe PublicIPRef (FQID only).
// Use publicIP.Ref() for resource chaining.
func (b *VirtualMachineBuilder) WithPublicIP(ref PublicIPRef) *VirtualMachineBuilder {
	b.publicIP = ref
	return b
}

// WithSecurityGroup accepts a type-safe SecurityGroupRef (FQID only).
// Use sg.Ref() for resource chaining.
func (b *VirtualMachineBuilder) WithSecurityGroup(ref SecurityGroupRef) *VirtualMachineBuilder {
	b.securityGroups = append(b.securityGroups, ref)
	return b
}

// WithSSHKey adds an SSH public key for authentication.
// WARNING: Ignored if WithCloudInit() is set. Configure SSH keys in your cloud-init script instead.
func (b *VirtualMachineBuilder) WithSSHKey(publicKey string) *VirtualMachineBuilder {
	b.sshKeys = append(b.sshKeys, publicKey)
	return b
}

// WithCloudInit sets custom cloud-init user data.
// WARNING: Replaces evroc's default cloud-init. WithSSHKey() will be IGNORED.
// You must configure SSH keys in your cloud-init script.
func (b *VirtualMachineBuilder) WithCloudInit(userData string) *VirtualMachineBuilder {
	b.cloudInitData = userData
	return b
}

// WithZone sets the availability zone for the VM.
func (b *VirtualMachineBuilder) WithZone(zone string) *VirtualMachineBuilder {
	b.zone = zone
	return b
}

// WithPlacementGroup accepts a type-safe PlacementGroupRef (FQID only).
// Use pg.Ref() for resource chaining.
func (b *VirtualMachineBuilder) WithPlacementGroup(ref PlacementGroupRef) *VirtualMachineBuilder {
	b.placementGroup = ref
	return b
}

// WithRunning sets whether the VM should be running after creation.
func (b *VirtualMachineBuilder) WithRunning(running bool) *VirtualMachineBuilder {
	b.running = &running
	return b
}

// WithLabels sets user-defined labels for the VM.
func (b *VirtualMachineBuilder) WithLabels(labels map[string]string) *VirtualMachineBuilder {
	b.labels = labels
	return b
}

// Build creates the VirtualMachineRequest structure ready for the Create API call.
func (b *VirtualMachineBuilder) Build() *compute.VirtualMachineRequest {
	vmReq := &compute.VirtualMachineRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "VirtualMachine",
		Metadata: compute.RegionalMetadataRequest{
			Id: b.id,
		},
		Spec: compute.VirtualMachineSpec{
			ComputeProfileRef: b.vmSize,
			Running:           b.running,
			Placement:         compute.VirtualMachineSpecPlacement{}, // Always required, non-pointer
		},
	}

	// Build disk references
	if len(b.diskRefs) > 0 {
		disks := make([]compute.VirtualMachineSpecDisksItem, len(b.diskRefs))
		for i, diskRef := range b.diskRefs {
			disks[i].DiskRef = diskRef.ref.String()
			if diskRef.bootFrom {
				bootFrom := true
				disks[i].BootFrom = &bootFrom
			}
		}
		vmReq.Spec.Disks = &disks
	}

	// Add networking configuration if specified
	if b.publicIP != "" || len(b.securityGroups) > 0 {
		vmReq.Spec.Networking = &compute.VirtualMachineSpecNetworking{}

		if b.publicIP != "" {
			publicIPStr := b.publicIP.String()
			vmReq.Spec.Networking.PublicIPv4Address = &struct {
				Static *compute.VirtualMachineSpecNetworkingStatic `json:"static,omitempty"`
			}{
				Static: &compute.VirtualMachineSpecNetworkingStatic{
					PublicIPRef: &publicIPStr,
				},
			}
		}

		if len(b.securityGroups) > 0 {
			sgRefs := make([]string, len(b.securityGroups))
			for i, ref := range b.securityGroups {
				sgRefs[i] = ref.String()
			}
			vmReq.Spec.Networking.SecurityGroupSettings = &struct {
				SecurityGroupMemberRefs *[]string `json:"securityGroupMemberRefs,omitempty"`
			}{
				SecurityGroupMemberRefs: &sgRefs,
			}
		}
	}

	// Warn if both SSH keys and cloud-init are set (SSH keys will be ignored)
	if b.cloudInitData != "" && len(b.sshKeys) > 0 {
		log.Println("WARNING: WithSSHKey() ignored when WithCloudInit() is set. Configure SSH keys in your cloud-init script.")
	}

	// Add OS settings if SSH keys or cloud-init data specified
	if len(b.sshKeys) > 0 || b.cloudInitData != "" {
		vmReq.Spec.OsSettings = &compute.VirtualMachineSpecOsSettings{}

		if b.cloudInitData != "" {
			vmReq.Spec.OsSettings.CloudInitUserData = &b.cloudInitData
		}

		if len(b.sshKeys) > 0 {
			authorizedKeys := make([]compute.VirtualMachineSpecOsSettingsAuthorizedKeysItem, len(b.sshKeys))
			for i, key := range b.sshKeys {
				authorizedKeys[i].Value = key
			}
			vmReq.Spec.OsSettings.Ssh = &struct {
				AuthorizedKeys *[]compute.VirtualMachineSpecOsSettingsAuthorizedKeysItem `json:"authorizedKeys,omitempty"`
			}{
				AuthorizedKeys: &authorizedKeys,
			}
		}
	}

	// Set placement zone (mandatory)
	vmReq.Spec.Placement.Zone = &b.zone

	// Add placement group if specified
	if b.placementGroup != "" {
		pgStr := b.placementGroup.String()
		vmReq.Spec.Placement.PlacementGroupRef = &pgStr
	}

	// Add labels if specified
	if len(b.labels) > 0 {
		userLabels := compute.UserLabels(b.labels)
		vmReq.Metadata.UserLabels = &userLabels
	}

	return vmReq
}

// Create is a convenience method that builds and creates the VM in one call.
func (b *VirtualMachineBuilder) Create(ctx context.Context, client *VirtualMachinesService) (*compute.VirtualMachine, error) {
	vmReq := b.Build()
	return client.Create(ctx, vmReq)
}

// ============================================================================
// HotswapDiskAttachment Builder
// ============================================================================

// HotswapDiskAttachmentBuilder provides a fluent interface for creating HotswapDiskAttachment resources.
type HotswapDiskAttachmentBuilder struct {
	id      string
	diskRef DiskRef
	vmRef   VMRef
	labels  map[string]string
}

// NewHotswapDiskAttachmentBuilder creates a new builder for HotswapDiskAttachment.
// Use client.Compute().VMRef(name) and client.Compute().DiskRef(name) to construct refs from names,
// or vm.Ref() and disk.Ref() to get refs from resource objects.
func NewHotswapDiskAttachmentBuilder(id string, vmRef VMRef, diskRef DiskRef) *HotswapDiskAttachmentBuilder {
	return &HotswapDiskAttachmentBuilder{
		id:      id,
		vmRef:   vmRef,
		diskRef: diskRef,
	}
}

// WithLabels sets user-defined labels for the hotswap disk attachment.
func (b *HotswapDiskAttachmentBuilder) WithLabels(labels map[string]string) *HotswapDiskAttachmentBuilder {
	b.labels = labels
	return b
}

// Build creates the HotswapDiskAttachmentRequest structure.
// Resource references use short IDs - the service layer will resolve them to Fully Qualified IDs (FQIDs).
func (b *HotswapDiskAttachmentBuilder) Build() *compute.HotswapDiskAttachmentRequest {
	req := &compute.HotswapDiskAttachmentRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "HotswapDiskAttachment",
		Metadata: compute.RegionalMetadataRequest{
			Id: b.id,
		},
		Spec: compute.HotswapDiskAttachmentSpec{
			VirtualMachineRef: b.vmRef.String(),
			DiskRef:           b.diskRef.String(),
		},
	}

	// Add labels if specified
	if len(b.labels) > 0 {
		userLabels := compute.UserLabels(b.labels)
		req.Metadata.UserLabels = &userLabels
	}

	return req
}

// Create is a convenience method that builds and creates the attachment in one call.
func (b *HotswapDiskAttachmentBuilder) Create(ctx context.Context, client *HotswapDiskAttachmentsService) (*compute.HotswapDiskAttachment, error) {
	req := b.Build()
	return client.Create(ctx, req)
}

// ============================================================================
// PlacementGroup Builder
// ============================================================================

// PlacementGroupBuilder provides a fluent interface for creating PlacementGroup resources.
type PlacementGroupBuilder struct {
	id       string
	strategy string
	zone     string
	labels   map[string]string
}

// NewPlacementGroupBuilder creates a new builder for PlacementGroup.
func NewPlacementGroupBuilder(id string, strategy string) *PlacementGroupBuilder {
	return &PlacementGroupBuilder{
		id:       id,
		strategy: strategy,
	}
}

// WithZone sets the zone for the placement group.
func (b *PlacementGroupBuilder) WithZone(zone string) *PlacementGroupBuilder {
	b.zone = zone
	return b
}

// WithLabels sets user-defined labels for the placement group.
func (b *PlacementGroupBuilder) WithLabels(labels map[string]string) *PlacementGroupBuilder {
	b.labels = labels
	return b
}

// Build creates the PlacementGroupRequest structure.
func (b *PlacementGroupBuilder) Build() *compute.PlacementGroupRequest {
	req := &compute.PlacementGroupRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "PlacementGroup",
		Metadata: compute.RegionalMetadataRequest{
			Id: b.id,
		},
		Spec: compute.PlacementGroupSpec{
			Strategy: compute.PlacementGroupSpecStrategy{
				Type: compute.PlacementGroupSpecStrategyType(b.strategy),
			},
			Placement: compute.PlacementGroupSpecPlacement{}, // Always required, non-pointer
		},
	}

	// Set placement zone (mandatory)
	req.Spec.Placement.Zone = &b.zone

	// Add labels if specified
	if len(b.labels) > 0 {
		userLabels := compute.UserLabels(b.labels)
		req.Metadata.UserLabels = &userLabels
	}

	return req
}

// Create is a convenience method that builds and creates the placement group in one call.
func (b *PlacementGroupBuilder) Create(ctx context.Context, client *PlacementGroupsService) (*compute.PlacementGroup, error) {
	req := b.Build()
	return client.Create(ctx, req)
}
