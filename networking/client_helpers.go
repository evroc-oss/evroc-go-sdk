// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"fmt"

	networkingtypes "github.com/evroc-oss/evroc-go-sdk/types/networking"
)

// PublicIPRef creates a PublicIPRef from a name using the client's project and region context.
// This is a convenience method for constructing FQIDs from names within the current context.
func (c *Client) PublicIPRef(name string) networkingtypes.PublicIPRef {
	return networkingtypes.PublicIPRef(fmt.Sprintf("/networking/projects/%s/regions/%s/publicIPs/%s",
		c.parent.DefaultProject(),
		c.parent.DefaultRegion(),
		name))
}

// VPCRef creates a VPC resource reference from a name using the client's project and region context.
func (c *Client) VPCRef(name string) string {
	return fmt.Sprintf("/networking/projects/%s/regions/%s/virtualPrivateClouds/%s",
		c.parent.DefaultProject(),
		c.parent.DefaultRegion(),
		name)
}

// DefaultVPCRef returns the reference to the default VPC for the client's region.
func (c *Client) DefaultVPCRef() string {
	return c.VPCRef("default-" + c.parent.DefaultRegion())
}

// SecurityGroupRef creates a SecurityGroupRef from a name using the client's project and region context.
// This is a convenience method for constructing FQIDs from names within the current context.
func (c *Client) SecurityGroupRef(name string) networkingtypes.SecurityGroupRef {
	return networkingtypes.SecurityGroupRef(fmt.Sprintf("/networking/projects/%s/regions/%s/securityGroups/%s",
		c.parent.DefaultProject(),
		c.parent.DefaultRegion(),
		name))
}
