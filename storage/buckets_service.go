// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"errors"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
)

// WaitForDeleted waits for a bucket to be fully deleted
func (s *BucketsService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration) error {
	config := rest.WaiterConfig{
		Timeout:         timeout,
		InitialInterval: 2 * time.Second,
		MaxInterval:     10 * time.Second,
		Multiplier:      1.5,
	}

	return rest.WaitFor(ctx, config, func() (bool, error) {
		_, err := s.Get(ctx, name)
		if errors.Is(err, rest.ErrNotFound) {
			// Resource deleted successfully
			return true, nil
		}
		// Still exists, keep polling
		return false, nil
	})
}
