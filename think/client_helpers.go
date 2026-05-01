// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package think

import (
	"context"
	"path"
)

// Start requests that an instance is started. It is a no-op on an already running instance.
func (s *InstancesService) Start(ctx context.Context, name string) error {
	resourcePath := path.Join(s.client.path.ResourcePath(
		s.client.parent.DefaultProject(),
		s.client.parent.DefaultRegion(),
		resourceInstances,
		name), "start")
	_, err := s.client.rest.Post(ctx, resourcePath, nil)
	return err
}

// Stop requests that an instance is stopped. It is a no-op on an already stopped instance.
func (s *InstancesService) Stop(ctx context.Context, name string) error {
	resourcePath := path.Join(s.client.path.ResourcePath(
		s.client.parent.DefaultProject(),
		s.client.parent.DefaultRegion(),
		resourceInstances,
		name), "stop")
	_, err := s.client.rest.Post(ctx, resourcePath, nil)
	return err
}
