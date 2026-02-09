// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package storage

import (
	"context"
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
		if err != nil {
			// Resource no longer exists - this is what we want
			return true, nil
		}
		return false, nil
	})
}
