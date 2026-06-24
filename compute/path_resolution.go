// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"fmt"
	"strings"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	compute "github.com/evroc-oss/evroc-go-sdk/types/compute"
)

// pathResolver provides helper methods for resolving resource references.
type pathResolver struct {
	project        string
	region         string
	computePath    rest.ServicePath
	networkingPath rest.ServicePath
}

// newPathResolver creates a new path resolver with the given context.
func (s *VirtualMachinesService) newPathResolver() *pathResolver {
	return &pathResolver{
		project:        s.client.parent.DefaultProject(),
		region:         s.client.parent.DefaultRegion(),
		computePath:    s.client.path,
		networkingPath: rest.NewServicePath("networking", apiVersion),
	}
}

// resolve resolves a short ID to a fully qualified ID using the given service and resource type.
func (r *pathResolver) resolve(ref string, path rest.ServicePath, resourceType string) string {
	if strings.HasPrefix(ref, "/") {
		return ref
	}
	// Build resource reference path WITHOUT API version
	// Format: /{service}/projects/{project}/regions/{region}/{resourceType}/{name}
	return fmt.Sprintf("/%s/projects/%s/regions/%s/%s/%s",
		path.Service(), r.project, r.region, resourceType, ref)
}

// resolvePtr resolves a pointer to a string reference.
func (r *pathResolver) resolvePtr(ref *string, path rest.ServicePath, resourceType string) {
	if ref != nil && !strings.HasPrefix(*ref, "/") {
		// Build resource reference path WITHOUT API version
		resolved := fmt.Sprintf("/%s/projects/%s/regions/%s/%s/%s",
			path.Service(), r.project, r.region, resourceType, *ref)
		*ref = resolved
	}
}

// resolveVirtualMachinesPaths resolves short IDs to fully qualified IDs.
func (s *VirtualMachinesService) resolveVirtualMachinesPaths(req *compute.VirtualMachineRequest) *compute.VirtualMachineRequest {
	resolved := *req
	r := s.newPathResolver()

	// Resolve disk references
	if resolved.Spec.Disks != nil {
		disks := make([]compute.VirtualMachineSpecDisksItem, len(*resolved.Spec.Disks))
		for i, disk := range *resolved.Spec.Disks {
			disks[i] = disk
			disks[i].DiskRef = r.resolve(disk.DiskRef, r.computePath, resourceDisks)
		}
		resolved.Spec.Disks = &disks
	}

	// Resolve networking references
	if resolved.Spec.Networking != nil {
		if resolved.Spec.Networking.PublicIPv4Address != nil &&
			resolved.Spec.Networking.PublicIPv4Address.Static != nil {
			r.resolvePtr(resolved.Spec.Networking.PublicIPv4Address.Static.PublicIPRef, r.networkingPath, "publicIPs")
		}

		if resolved.Spec.Networking.SecurityGroupSettings != nil &&
			resolved.Spec.Networking.SecurityGroupSettings.SecurityGroupMemberRefs != nil {
			sgRefs := *resolved.Spec.Networking.SecurityGroupSettings.SecurityGroupMemberRefs
			for i := range sgRefs {
				sgRefs[i] = r.resolve(sgRefs[i], r.networkingPath, "securityGroups")
			}
		}
	}

	// Resolve placement group reference
	r.resolvePtr(resolved.Spec.Placement.PlacementGroupRef, r.computePath, resourcePlacementGroups)

	// Resolve compute profile reference (global resource)
	if !strings.HasPrefix(resolved.Spec.ComputeProfileRef, "/") {
		resolved.Spec.ComputeProfileRef = fmt.Sprintf("/compute/global/computeProfiles/%s", resolved.Spec.ComputeProfileRef)
	}

	return &resolved
}

// resolveHotswapDiskAttachmentsPaths resolves short IDs to fully qualified IDs.
func (s *HotswapDiskAttachmentsService) resolveHotswapDiskAttachmentsPaths(req *compute.HotswapDiskAttachmentRequest) *compute.HotswapDiskAttachmentRequest {
	resolved := *req
	r := &pathResolver{
		project:     s.client.parent.DefaultProject(),
		region:      s.client.parent.DefaultRegion(),
		computePath: s.client.path,
	}

	resolved.Spec.VirtualMachineRef = r.resolve(resolved.Spec.VirtualMachineRef, r.computePath, resourceVirtualMachines)
	resolved.Spec.DiskRef = r.resolve(resolved.Spec.DiskRef, r.computePath, resourceDisks)

	return &resolved
}

// resolveSecurityGroupPath resolves a short security group ID to a fully qualified ID.
func (s *VirtualMachinesService) resolveSecurityGroupPath(sgName string) string {
	return s.newPathResolver().resolve(sgName, rest.NewServicePath("networking", apiVersion), "securityGroups")
}

// resolvePublicIPPath resolves a short public IP ID to a fully qualified ID.
func (s *VirtualMachinesService) resolvePublicIPPath(publicIPName string) string {
	return s.newPathResolver().resolve(publicIPName, rest.NewServicePath("networking", apiVersion), "publicIPs")
}

// resolvePlacementGroupPath resolves a short placement group ID to a fully qualified ID.
func (s *VirtualMachinesService) resolvePlacementGroupPath(pgName string) string {
	return s.newPathResolver().resolve(pgName, s.client.path, resourcePlacementGroups)
}
