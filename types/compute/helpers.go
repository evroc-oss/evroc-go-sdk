// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import "fmt"

// DiskRef is a type-safe reference to a disk resource.
type DiskRef string

// String returns the string representation of the DiskRef.
func (r DiskRef) String() string {
	return string(r)
}

// VMRef is a type-safe reference to a virtual machine resource.
type VMRef string

// String returns the string representation of the VMRef.
func (r VMRef) String() string {
	return string(r)
}

// PlacementGroupRef is a type-safe reference to a placement group resource.
type PlacementGroupRef string

// String returns the string representation of the PlacementGroupRef.
func (r PlacementGroupRef) String() string {
	return string(r)
}

// Ref returns the fully qualified reference as a type-safe DiskRef for use with builders.
func (d *Disk) Ref() DiskRef {
	if d.Metadata.Project == nil || d.Metadata.Region == nil || d.Metadata.Id == "" {
		return ""
	}
	fqid := fmt.Sprintf("/compute/projects/%s/regions/%s/disks/%s",
		*d.Metadata.Project, *d.Metadata.Region, d.Metadata.Id)
	return DiskRef(fqid)
}

// Ref returns the fully qualified reference as a type-safe VMRef for use with builders.
func (vm *VirtualMachine) Ref() VMRef {
	if vm.Metadata.Project == nil || vm.Metadata.Region == nil || vm.Metadata.Id == "" {
		return ""
	}
	fqid := fmt.Sprintf("/compute/projects/%s/regions/%s/virtualMachines/%s",
		*vm.Metadata.Project, *vm.Metadata.Region, vm.Metadata.Id)
	return VMRef(fqid)
}

// Ref returns the fully qualified reference as a type-safe PlacementGroupRef for use with builders.
func (pg *PlacementGroup) Ref() PlacementGroupRef {
	if pg.Metadata.Project == nil || pg.Metadata.Region == nil || pg.Metadata.Id == "" {
		return ""
	}
	fqid := fmt.Sprintf("/compute/projects/%s/regions/%s/placementGroups/%s",
		*pg.Metadata.Project, *pg.Metadata.Region, pg.Metadata.Id)
	return PlacementGroupRef(fqid)
}

// Ref returns the fully qualified reference path for a HotswapDiskAttachment.
func (hda *HotswapDiskAttachment) Ref() string {
	if hda.Metadata.Project == nil || hda.Metadata.Region == nil || hda.Metadata.Id == "" {
		return ""
	}
	return fmt.Sprintf("/compute/projects/%s/regions/%s/hotswapDiskAttachments/%s",
		*hda.Metadata.Project, *hda.Metadata.Region, hda.Metadata.Id)
}
