// Package compute provides builder patterns for compute resources.
package compute

import (
	"context"

	compute "github.com/evroc-oss/evroc-go-sdk/types/compute"
)

// DiskBuilder provides a fluent interface for creating Disk resources.
type DiskBuilder struct {
	name         string
	image        string
	sizeAmount   int32
	sizeUnit     string
	storageClass compute.DiskSpecDiskStorageClassName
	zone         string
	labels       map[string]string
}

// NewDiskBuilder creates a new DiskBuilder with sensible defaults.
func NewDiskBuilder(name string) *DiskBuilder {
	return &DiskBuilder{
		name:         name,
		storageClass: compute.Persistent, // Default to persistent storage
		sizeAmount:   10,                 // Default 10GB
		sizeUnit:     "GB",
	}
}

// WithImage sets the disk image (e.g., "ubuntu-minimal.24-04.1", "ubuntu.24-04.1", "rocky.10-0.1").
func (b *DiskBuilder) WithImage(image string) *DiskBuilder {
	b.image = image
	return b
}

// WithSize sets the disk size.
func (b *DiskBuilder) WithSize(amount int32, unit string) *DiskBuilder {
	b.sizeAmount = amount
	b.sizeUnit = unit
	return b
}

// WithSizeGB is a convenience method to set size in GB.
func (b *DiskBuilder) WithSizeGB(sizeGB int32) *DiskBuilder {
	return b.WithSize(sizeGB, "GB")
}

// WithStorageClass sets the storage class (currently only "persistent" is supported).
func (b *DiskBuilder) WithStorageClass(storageClass compute.DiskSpecDiskStorageClassName) *DiskBuilder {
	b.storageClass = storageClass
	return b
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
		ApiVersion: "compute/v1alpha2",
		Kind:       "Disk",
		Metadata: compute.RegionalMetadataRequest{
			Name: &b.name,
		},
		Spec: compute.DiskSpec{
			DiskStorageClass: compute.DiskSpecDiskStorageClass{
				Name: b.storageClass,
			},
		},
	}

	// Add optional image
	if b.image != "" {
		diskReq.Spec.DiskImage = &compute.DiskSpecDiskImage{
			DiskImageRef: struct {
				Name string `json:"name"`
			}{
				Name: b.image,
			},
		}
	}

	// Add custom disk size if specified
	if b.sizeAmount > 0 {
		diskReq.Spec.DiskSize = &compute.DiskSpecDiskSize{
			Amount: b.sizeAmount,
			Unit:   compute.DiskSpecDiskSizeUnit(b.sizeUnit),
		}
	}

	// Add placement zone if specified
	if b.zone != "" {
		diskReq.Spec.Placement = &compute.DiskSpecPlacement{
			Zone: &b.zone,
		}
	}

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
	name           string
	diskRefs       []diskRef
	vmSize         string
	publicIP       string
	securityGroups []string
	sshKeys        []string
	cloudInitData  string
	zone           string
	placementGroup string
	running        *bool
	labels         map[string]string
}

type diskRef struct {
	name     string
	bootFrom bool
}

// NewVirtualMachineBuilder creates a new VirtualMachineBuilder with sensible defaults.
func NewVirtualMachineBuilder(name string) *VirtualMachineBuilder {
	running := true
	return &VirtualMachineBuilder{
		name:     name,
		vmSize:   "a1a.xs", // Default to smallest size
		diskRefs: []diskRef{},
		running:  &running,
	}
}

// WithBootDisk adds a boot disk reference.
func (b *VirtualMachineBuilder) WithBootDisk(diskName string) *VirtualMachineBuilder {
	b.diskRefs = append(b.diskRefs, diskRef{
		name:     diskName,
		bootFrom: true,
	})
	return b
}

// WithDataDisk adds a data disk reference.
func (b *VirtualMachineBuilder) WithDataDisk(diskName string) *VirtualMachineBuilder {
	b.diskRefs = append(b.diskRefs, diskRef{
		name:     diskName,
		bootFrom: false,
	})
	return b
}

// WithSize sets the VM size (e.g., "a1a.xs", "c1a.m", "m1a.l").
func (b *VirtualMachineBuilder) WithSize(size string) *VirtualMachineBuilder {
	b.vmSize = size
	return b
}

// WithPublicIP attaches a public IP to the VM.
func (b *VirtualMachineBuilder) WithPublicIP(publicIPRef string) *VirtualMachineBuilder {
	b.publicIP = publicIPRef
	return b
}

// WithSecurityGroup adds a security group to the VM.
func (b *VirtualMachineBuilder) WithSecurityGroup(sgName string) *VirtualMachineBuilder {
	b.securityGroups = append(b.securityGroups, sgName)
	return b
}

// WithSSHKey adds an SSH public key for authentication.
func (b *VirtualMachineBuilder) WithSSHKey(publicKey string) *VirtualMachineBuilder {
	b.sshKeys = append(b.sshKeys, publicKey)
	return b
}

// WithCloudInit sets cloud-init user data.
func (b *VirtualMachineBuilder) WithCloudInit(userData string) *VirtualMachineBuilder {
	b.cloudInitData = userData
	return b
}

// WithZone sets the availability zone for the VM.
func (b *VirtualMachineBuilder) WithZone(zone string) *VirtualMachineBuilder {
	b.zone = zone
	return b
}

// WithPlacementGroup sets the placement group for the VM.
func (b *VirtualMachineBuilder) WithPlacementGroup(pgName string) *VirtualMachineBuilder {
	b.placementGroup = pgName
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
		ApiVersion: "compute/v1alpha2",
		Kind:       "VirtualMachine",
		Metadata: compute.RegionalMetadataRequest{
			Name: &b.name,
		},
		Spec: compute.VirtualMachineSpec{
			VmVirtualResourcesRef: compute.VirtualMachineSpecVmVirtualResourcesRef{
				VmVirtualResourcesRefName: b.vmSize,
			},
			Running: b.running,
		},
	}

	// Build disk references
	if len(b.diskRefs) > 0 {
		vmReq.Spec.DiskRefs = make([]compute.VirtualMachineSpecDiskRefsItem, len(b.diskRefs))
		for i, ref := range b.diskRefs {
			vmReq.Spec.DiskRefs[i].Name = ref.name
			if ref.bootFrom {
				bootFrom := true
				vmReq.Spec.DiskRefs[i].BootFrom = &bootFrom
			}
		}
	}

	// Add networking configuration if specified
	if b.publicIP != "" || len(b.securityGroups) > 0 {
		vmReq.Spec.Networking = &compute.VirtualMachineSpecNetworking{}

		if b.publicIP != "" {
			vmReq.Spec.Networking.PublicIPv4Address = &struct {
				Static *compute.VirtualMachineSpecNetworkingStatic `json:"static,omitempty"`
			}{
				Static: &compute.VirtualMachineSpecNetworkingStatic{
					PublicIPRef: &b.publicIP,
				},
			}
		}

		if len(b.securityGroups) > 0 {
			sgMemberships := make([]compute.VirtualMachineSpecNetworkingSecurityGroupMembershipsItem, len(b.securityGroups))
			for i, sg := range b.securityGroups {
				sgName := sg
				sgMemberships[i].Name = &sgName
			}
			vmReq.Spec.Networking.SecurityGroups = &struct {
				SecurityGroupMemberships *[]compute.VirtualMachineSpecNetworkingSecurityGroupMembershipsItem `json:"securityGroupMemberships,omitempty"`
			}{
				SecurityGroupMemberships: &sgMemberships,
			}
		}
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
				keyValue := key
				authorizedKeys[i].Value = &keyValue
			}
			vmReq.Spec.OsSettings.Ssh = &struct {
				AuthorizedKeys *[]compute.VirtualMachineSpecOsSettingsAuthorizedKeysItem `json:"authorizedKeys,omitempty"`
			}{
				AuthorizedKeys: &authorizedKeys,
			}
		}
	}

	// Add placement if zone specified
	if b.zone != "" {
		vmReq.Spec.Placement = &compute.VirtualMachineSpecPlacement{
			Zone: &b.zone,
		}
	}

	// Add placement group if specified
	if b.placementGroup != "" {
		vmReq.Spec.PlacementGroup = &b.placementGroup
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
	name    string
	diskRef string
	vmRef   string
	labels  map[string]string
}

// NewHotswapDiskAttachmentBuilder creates a new builder for HotswapDiskAttachment.
func NewHotswapDiskAttachmentBuilder(name string, vmRef string, diskRef string) *HotswapDiskAttachmentBuilder {
	return &HotswapDiskAttachmentBuilder{
		name:    name,
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
func (b *HotswapDiskAttachmentBuilder) Build() *compute.HotswapDiskAttachmentRequest {
	req := &compute.HotswapDiskAttachmentRequest{
		ApiVersion: "compute/v1alpha2",
		Kind:       "HotswapDiskAttachment",
		Metadata: compute.RegionalMetadataRequest{
			Name: &b.name,
		},
		Spec: compute.HotswapDiskAttachmentSpec{
			VmRef:   b.vmRef,
			DiskRef: b.diskRef,
		},
	}

	// Add labels if specified
	if len(b.labels) > 0 {
		userLabels := compute.UserLabels(b.labels); req.Metadata.UserLabels = &userLabels
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
	name     string
	strategy string
	zone     string
	labels   map[string]string
}

// NewPlacementGroupBuilder creates a new builder for PlacementGroup
// strategy can be "spread".
func NewPlacementGroupBuilder(name string, strategy string) *PlacementGroupBuilder {
	return &PlacementGroupBuilder{
		name:     name,
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
		ApiVersion: "compute/v1alpha2",
		Kind:       "PlacementGroup",
		Metadata: compute.RegionalMetadataRequest{
			Name: &b.name,
		},
		Spec: compute.PlacementGroupSpec{
			Strategy: compute.PlacementGroupSpecStrategy{
				Type: compute.PlacementGroupSpecStrategyType(b.strategy),
			},
		},
	}

	if b.zone != "" {
		req.Spec.Placement = &compute.PlacementGroupSpecPlacement{
			Zone: &b.zone,
		}
	}

	// Add labels if specified
	if len(b.labels) > 0 {
		userLabels := compute.UserLabels(b.labels); req.Metadata.UserLabels = &userLabels
	}

	return req
}

// Create is a convenience method that builds and creates the placement group in one call.
func (b *PlacementGroupBuilder) Create(ctx context.Context, client *PlacementGroupsService) (*compute.PlacementGroup, error) {
	req := b.Build()
	return client.Create(ctx, req)
}
