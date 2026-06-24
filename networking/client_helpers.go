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

// SecurityGroupRef creates a SecurityGroupRef from a name using the client's project and region context.
// This is a convenience method for constructing FQIDs from names within the current context.
func (c *Client) SecurityGroupRef(name string) networkingtypes.SecurityGroupRef {
	return networkingtypes.SecurityGroupRef(fmt.Sprintf("/networking/projects/%s/regions/%s/securityGroups/%s",
		c.parent.DefaultProject(),
		c.parent.DefaultRegion(),
		name))
}
