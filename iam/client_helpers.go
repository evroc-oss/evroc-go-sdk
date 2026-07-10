// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package iam

import "fmt"

// ServiceAccountRef creates a fully-qualified service account reference from a name.
func (c *Client) ServiceAccountRef(name string) string {
	return fmt.Sprintf("/iam/projects/%s/serviceAccounts/%s",
		c.parent.DefaultProject(), name)
}
