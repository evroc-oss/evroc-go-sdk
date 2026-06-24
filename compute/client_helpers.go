// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import "fmt"

// DiskRef creates a DiskRef from a disk name using the client's project and region context.
// This is a convenience method for constructing FQIDs from names within the current context.
func (c *Client) DiskRef(name string) DiskRef {
	return DiskRef(fmt.Sprintf("/compute/projects/%s/regions/%s/disks/%s",
		c.parent.DefaultProject(),
		c.parent.DefaultRegion(),
		name))
}

// VMRef creates a VMRef from a VM name using the client's project and region context.
// This is a convenience method for constructing FQIDs from names within the current context.
func (c *Client) VMRef(name string) VMRef {
	return VMRef(fmt.Sprintf("/compute/projects/%s/regions/%s/virtualMachines/%s",
		c.parent.DefaultProject(),
		c.parent.DefaultRegion(),
		name))
}

// PlacementGroupRef creates a PlacementGroupRef from a name using the client's project and region context.
// This is a convenience method for constructing FQIDs from names within the current context.
func (c *Client) PlacementGroupRef(name string) PlacementGroupRef {
	return PlacementGroupRef(fmt.Sprintf("/compute/projects/%s/regions/%s/placementGroups/%s",
		c.parent.DefaultProject(),
		c.parent.DefaultRegion(),
		name))
}
