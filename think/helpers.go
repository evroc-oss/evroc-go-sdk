// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package think

import (
	"fmt"

	"github.com/evroc-oss/evroc-go-sdk/types/think"
)

// ============================================================================
// Status Helpers
// ============================================================================

// GetInstancePhase returns the current phase of the Instance, or empty string if unknown.
func GetInstancePhase(instance *think.Instance) think.InstanceStatusPhase {
	if instance == nil || instance.Status.Phase == nil {
		return ""
	}
	return *instance.Status.Phase
}

// IsInstanceRunning returns true if the Instance is in the Running phase.
func IsInstanceRunning(instance *think.Instance) bool {
	return GetInstancePhase(instance) == think.Running
}

// IsInstanceStopped returns true if the Instance is in the Stopped phase.
func IsInstanceStopped(instance *think.Instance) bool {
	return GetInstancePhase(instance) == think.Stopped
}

// IsInstanceFailed returns true if the Instance is in the Failed phase.
func IsInstanceFailed(instance *think.Instance) bool {
	return GetInstancePhase(instance) == think.Failed
}

// ============================================================================
// Ref Helpers
// ============================================================================

// InstanceRef creates a fully qualified instance reference from a name
// using the client's project and region context.
func (c *Client) InstanceRef(name string) string {
	return fmt.Sprintf("/think/projects/%s/regions/%s/instances/%s",
		c.parent.DefaultProject(),
		c.parent.DefaultRegion(),
		name)
}
