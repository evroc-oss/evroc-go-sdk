// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"context"
	"path"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/types/compute"
)

func (s *VirtualMachinesService) actionPath(name string, action ...string) string {
	return path.Join(append([]string{s.client.path.ResourcePath(
		s.client.parent.DefaultProject(),
		s.client.parent.DefaultRegion(),
		resourceVirtualMachines,
		name)}, action...)...)
}

// ConsoleLogs returns the last 1,000 lines of a VirtualMachine's console logs.
func (s *VirtualMachinesService) ConsoleLogs(ctx context.Context, name string) (*compute.LogsConsole, error) {
	return rest.GetResource[*compute.LogsConsole](ctx, s.client.rest, s.actionPath(name, "logs", "console"))
}
