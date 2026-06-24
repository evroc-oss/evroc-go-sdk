// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"context"
	"fmt"
	"strings"

	compute "github.com/evroc-oss/evroc-go-sdk/types/compute"
)

// Update Builders
//
// Update builders provide a fluent interface for modifying existing resources,
// eliminating the verbose get-modify-update pattern:
//
//	// Without update builders (verbose):
//	vm, err := service.Get(ctx, "my-vm")
//	if err != nil { return err }
//	running := false
//	vm.Spec.Running = &running
//	_, err = service.Patch(ctx, "my-vm", vm)
//
//	// With update builders (concise):
//	_, err := UpdateVM("my-vm", service).
//		Stop().
//		Apply(ctx)
//
// All update builders follow the same pattern:
//  1. Create builder with resource name and service
//  2. Chain modification methods
//  3. Call Apply(ctx) to execute the update
//
// Available update builders:
//   - UpdateVM: Modify VM state, size, labels, security groups, public IPs, placement groups
//   - UpdateDisk: Modify labels (disk resize not supported by platform)

// VirtualMachineUpdateBuilder provides a fluent interface for updating VirtualMachine resources.
//
// Example:
//
//	updated, err := UpdateVM("my-vm", vmService).
//		Stop().
//		AddLabel("maintenance", "true").
//		Apply(ctx)
type VirtualMachineUpdateBuilder struct {
	name                 string
	service              *VirtualMachinesService
	updates              map[string]interface{}
	addLabels            map[string]string
	removeLabels         []string
	addSGs               []string
	removeSGs            []string
	running              *bool
	newSize              *string
	publicIP             *string
	removePublicIP       bool
	placementGroup       *string
	removePlacementGroup bool
}

// NewVirtualMachineUpdateBuilder creates a new builder for updating a VM.
func NewVirtualMachineUpdateBuilder(name string, service *VirtualMachinesService) *VirtualMachineUpdateBuilder {
	return &VirtualMachineUpdateBuilder{
		name:         name,
		service:      service,
		updates:      make(map[string]interface{}),
		addLabels:    make(map[string]string),
		removeLabels: []string{},
		addSGs:       []string{},
		removeSGs:    []string{},
	}
}

// Stop marks the VM to be stopped.
func (b *VirtualMachineUpdateBuilder) Stop() *VirtualMachineUpdateBuilder {
	running := false
	b.running = &running
	b.updates["running"] = false
	return b
}

// Start marks the VM to be started.
func (b *VirtualMachineUpdateBuilder) Start() *VirtualMachineUpdateBuilder {
	running := true
	b.running = &running
	b.updates["running"] = true
	return b
}

// AddLabel adds a label to the VM.
func (b *VirtualMachineUpdateBuilder) AddLabel(key, value string) *VirtualMachineUpdateBuilder {
	b.addLabels[key] = value
	b.updates["add_labels"] = b.addLabels
	return b
}

// RemoveLabel removes a label from the VM.
func (b *VirtualMachineUpdateBuilder) RemoveLabel(key string) *VirtualMachineUpdateBuilder {
	b.removeLabels = append(b.removeLabels, key)
	b.updates["remove_labels"] = b.removeLabels
	return b
}

// AddSecurityGroup adds a security group to the VM.
func (b *VirtualMachineUpdateBuilder) AddSecurityGroup(sgName string) *VirtualMachineUpdateBuilder {
	b.addSGs = append(b.addSGs, sgName)
	b.updates["add_sgs"] = b.addSGs
	return b
}

// RemoveSecurityGroup removes a security group from the VM.
func (b *VirtualMachineUpdateBuilder) RemoveSecurityGroup(sgName string) *VirtualMachineUpdateBuilder {
	b.removeSGs = append(b.removeSGs, sgName)
	b.updates["remove_sgs"] = b.removeSGs
	return b
}

// Resize changes the VM compute profile (size).
// VM must be stopped before changing compute profile.
func (b *VirtualMachineUpdateBuilder) Resize(newSize string) *VirtualMachineUpdateBuilder {
	b.newSize = &newSize
	b.updates["resize"] = newSize
	return b
}

// SetPublicIP changes the public IP attached to the VM.
// VM must be stopped before changing public IP.
func (b *VirtualMachineUpdateBuilder) SetPublicIP(publicIPRef string) *VirtualMachineUpdateBuilder {
	b.publicIP = &publicIPRef
	b.removePublicIP = false
	b.updates["public_ip"] = publicIPRef
	return b
}

// RemovePublicIP removes the public IP from the VM.
// VM must be stopped before removing public IP.
func (b *VirtualMachineUpdateBuilder) RemovePublicIP() *VirtualMachineUpdateBuilder {
	b.removePublicIP = true
	b.publicIP = nil
	b.updates["remove_public_ip"] = true
	return b
}

// SetPlacementGroup changes the placement group for the VM.
// VM must be stopped before changing placement group.
func (b *VirtualMachineUpdateBuilder) SetPlacementGroup(placementGroupRef string) *VirtualMachineUpdateBuilder {
	b.placementGroup = &placementGroupRef
	b.removePlacementGroup = false
	b.updates["placement_group"] = placementGroupRef
	return b
}

// RemovePlacementGroup removes the placement group from the VM.
// VM must be stopped before removing placement group.
func (b *VirtualMachineUpdateBuilder) RemovePlacementGroup() *VirtualMachineUpdateBuilder {
	b.removePlacementGroup = true
	b.placementGroup = nil
	b.updates["remove_placement_group"] = true
	return b
}

// Apply applies all pending updates to the VM.
func (b *VirtualMachineUpdateBuilder) Apply(ctx context.Context) (*compute.VirtualMachine, error) {
	if len(b.updates) == 0 {
		return nil, fmt.Errorf("no updates to apply")
	}

	// Fetch current VM state
	vm, err := b.service.Get(ctx, b.name)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VM: %w", err)
	}

	// Apply running state change
	if b.running != nil {
		vm.Spec.Running = b.running
	}

	// Apply size change
	if b.newSize != nil {
		// Resolve compute profile to fully qualified ID (global resource)
		if !strings.HasPrefix(*b.newSize, "/") {
			vm.Spec.ComputeProfileRef = fmt.Sprintf("/compute/global/computeProfiles/%s", *b.newSize)
		} else {
			vm.Spec.ComputeProfileRef = *b.newSize
		}
	}

	// Apply public IP changes
	if b.publicIP != nil || b.removePublicIP {
		if vm.Spec.Networking == nil {
			vm.Spec.Networking = &compute.VirtualMachineSpecNetworking{}
		}

		if b.removePublicIP {
			// Remove public IP
			vm.Spec.Networking.PublicIPv4Address = nil
		} else if b.publicIP != nil {
			// Set or change public IP
			publicIPPath := b.service.resolvePublicIPPath(*b.publicIP)
			vm.Spec.Networking.PublicIPv4Address = &struct {
				Static *compute.VirtualMachineSpecNetworkingStatic `json:"static,omitempty"`
			}{
				Static: &compute.VirtualMachineSpecNetworkingStatic{
					PublicIPRef: &publicIPPath,
				},
			}
		}
	}

	// Apply placement group changes
	if b.placementGroup != nil || b.removePlacementGroup {
		if b.removePlacementGroup {
			// Remove placement group
			vm.Spec.Placement.PlacementGroupRef = nil
		} else if b.placementGroup != nil {
			// Set or change placement group
			placementGroupPath := b.service.resolvePlacementGroupPath(*b.placementGroup)
			vm.Spec.Placement.PlacementGroupRef = &placementGroupPath
		}
	}

	// Apply label changes
	if len(b.addLabels) > 0 || len(b.removeLabels) > 0 {
		if vm.Metadata.UserLabels == nil {
			labels := make(compute.UserLabels)
			vm.Metadata.UserLabels = &labels
		}

		for k, v := range b.addLabels {
			(*vm.Metadata.UserLabels)[k] = v
		}

		for _, k := range b.removeLabels {
			delete(*vm.Metadata.UserLabels, k)
		}
	}

	// Apply security group changes
	if len(b.addSGs) > 0 || len(b.removeSGs) > 0 {
		if vm.Spec.Networking == nil {
			vm.Spec.Networking = &compute.VirtualMachineSpecNetworking{}
		}
		if vm.Spec.Networking.SecurityGroupSettings == nil {
			vm.Spec.Networking.SecurityGroupSettings = &struct {
				SecurityGroupMemberRefs *[]string `json:"securityGroupMemberRefs,omitempty"`
			}{
				SecurityGroupMemberRefs: &[]string{},
			}
		}

		// Build a map of existing SGs (already fully qualified IDs from API)
		sgMap := make(map[string]bool)
		if vm.Spec.Networking.SecurityGroupSettings.SecurityGroupMemberRefs != nil {
			for _, sg := range *vm.Spec.Networking.SecurityGroupSettings.SecurityGroupMemberRefs {
				sgMap[sg] = true
			}
		}

		// Add new SGs with fully qualified IDs
		for _, sgName := range b.addSGs {
			sgPath := b.service.resolveSecurityGroupPath(sgName)
			sgMap[sgPath] = true
		}

		// Remove SGs (convert short IDs to fully qualified IDs for removal)
		for _, sgName := range b.removeSGs {
			sgPath := b.service.resolveSecurityGroupPath(sgName)
			delete(sgMap, sgPath)
		}

		// Rebuild the SG list
		newSGs := make([]string, 0, len(sgMap))
		for sgPath := range sgMap {
			newSGs = append(newSGs, sgPath)
		}
		vm.Spec.Networking.SecurityGroupSettings.SecurityGroupMemberRefs = &newSGs
	}

	// Send the update
	updated, err := b.service.Patch(ctx, b.name, vm)
	if err != nil {
		return nil, fmt.Errorf("failed to update VM: %w", err)
	}

	return updated, nil
}

// DiskUpdateBuilder provides a fluent interface for updating Disk resources.
type DiskUpdateBuilder struct {
	name         string
	service      *DisksService
	updates      map[string]interface{}
	newSize      *int32
	addLabels    map[string]string
	removeLabels []string
}

// NewDiskUpdateBuilder creates a new builder for updating a disk.
func NewDiskUpdateBuilder(name string, service *DisksService) *DiskUpdateBuilder {
	return &DiskUpdateBuilder{
		name:         name,
		service:      service,
		updates:      make(map[string]interface{}),
		addLabels:    make(map[string]string),
		removeLabels: []string{},
	}
}

// ResizeGB sets a new size for the disk in GB.
func (b *DiskUpdateBuilder) ResizeGB(sizeGB int32) *DiskUpdateBuilder {
	b.newSize = &sizeGB
	b.updates["resize"] = sizeGB
	return b
}

// AddLabel adds a label to the disk.
func (b *DiskUpdateBuilder) AddLabel(key, value string) *DiskUpdateBuilder {
	b.addLabels[key] = value
	b.updates["add_labels"] = b.addLabels
	return b
}

// RemoveLabel removes a label from the disk.
func (b *DiskUpdateBuilder) RemoveLabel(key string) *DiskUpdateBuilder {
	b.removeLabels = append(b.removeLabels, key)
	b.updates["remove_labels"] = b.removeLabels
	return b
}

// Apply applies all pending updates to the disk.
func (b *DiskUpdateBuilder) Apply(ctx context.Context) (*compute.Disk, error) {
	if len(b.updates) == 0 {
		return nil, fmt.Errorf("no updates to apply")
	}

	// Fetch current disk state
	disk, err := b.service.Get(ctx, b.name)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch disk: %w", err)
	}

	// Apply size change
	if b.newSize != nil {
		if disk.Spec.DiskSize == nil {
			disk.Spec.DiskSize = &compute.DiskSpecDiskSize{}
		}
		disk.Spec.DiskSize.Amount = *b.newSize
		disk.Spec.DiskSize.Unit = "GB"
	}

	// Apply label changes
	if len(b.addLabels) > 0 || len(b.removeLabels) > 0 {
		if disk.Metadata.UserLabels == nil {
			labels := make(compute.UserLabels)
			disk.Metadata.UserLabels = &labels
		}

		for k, v := range b.addLabels {
			(*disk.Metadata.UserLabels)[k] = v
		}

		for _, k := range b.removeLabels {
			delete(*disk.Metadata.UserLabels, k)
		}
	}

	// Send the update
	updated, err := b.service.Patch(ctx, b.name, disk)
	if err != nil {
		return nil, fmt.Errorf("failed to update disk: %w", err)
	}

	return updated, nil
}

// Convenience functions for common update patterns

// UpdateVM creates an update builder for a VM.
func UpdateVM(name string, service *VirtualMachinesService) *VirtualMachineUpdateBuilder {
	return NewVirtualMachineUpdateBuilder(name, service)
}

// UpdateDisk creates an update builder for a disk.
func UpdateDisk(name string, service *DisksService) *DiskUpdateBuilder {
	return NewDiskUpdateBuilder(name, service)
}
