// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package loadbalancer

import (
	"context"
	"fmt"
)

// AddBackendRef adds a backend reference to a pool (idempotent).
// If the ref is already present, this is a no-op.
func (s *BackendPoolsService) AddBackendRef(ctx context.Context, poolName, backendRef string) error {
	pool, err := s.Get(ctx, poolName)
	if err != nil {
		return fmt.Errorf("get backend pool %s: %w", poolName, err)
	}

	var current []string
	if pool.Spec.BackendRefs != nil {
		current = *pool.Spec.BackendRefs
	}

	for _, ref := range current {
		if ref == backendRef {
			return nil
		}
	}

	updated := append(current, backendRef)
	pool.Spec.BackendRefs = &updated
	_, err = s.Patch(ctx, poolName, pool)
	return err
}

// RemoveBackendRef removes a backend reference from a pool (idempotent).
// If the ref is not present, this is a no-op.
func (s *BackendPoolsService) RemoveBackendRef(ctx context.Context, poolName, backendRef string) error {
	pool, err := s.Get(ctx, poolName)
	if err != nil {
		return fmt.Errorf("get backend pool %s: %w", poolName, err)
	}

	if pool.Spec.BackendRefs == nil {
		return nil
	}

	filtered := make([]string, 0, len(*pool.Spec.BackendRefs))
	for _, ref := range *pool.Spec.BackendRefs {
		if ref != backendRef {
			filtered = append(filtered, ref)
		}
	}

	if len(filtered) == len(*pool.Spec.BackendRefs) {
		return nil
	}

	pool.Spec.BackendRefs = &filtered
	_, err = s.Patch(ctx, poolName, pool)
	return err
}
