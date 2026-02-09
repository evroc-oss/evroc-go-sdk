// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package storage

import (
	"context"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
)

// WaitForCredentials waits for S3 credentials to be ready for a service account
func (s *BucketServiceAccountsService) WaitForCredentials(ctx context.Context, name string, timeout time.Duration) error {
	config := rest.WaiterConfig{
		Timeout:         timeout,
		InitialInterval: 2 * time.Second,
		MaxInterval:     10 * time.Second,
		Multiplier:      1.5,
	}

	return rest.WaitFor(ctx, config, func() (bool, error) {
		sa, err := s.Get(ctx, name)
		if err != nil {
			return false, err
		}
		return sa.Status.S3CredentialsSecretName != nil, nil
	})
}

// WaitForDeleted waits for a service account to be fully deleted
func (s *BucketServiceAccountsService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration) error {
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
