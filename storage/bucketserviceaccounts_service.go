// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"errors"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
)

// S3Credentials contains the S3-compatible access credentials for a bucket service account.
type S3Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
}

// WaitForCredentials waits for S3-compatible credentials to be ready and returns them.
// This eliminates the need to call Get() twice after waiting.
func (s *BucketServiceAccountsService) WaitForCredentials(ctx context.Context, name string, timeout time.Duration) (*S3Credentials, error) {
	config := rest.WaiterConfig{
		Timeout:         timeout,
		InitialInterval: 2 * time.Second,
		MaxInterval:     10 * time.Second,
		Multiplier:      1.5,
	}

	var result *S3Credentials
	err := rest.WaitFor(ctx, config, func() (bool, error) {
		sa, err := s.Get(ctx, name)
		if err != nil {
			// Transient errors during polling are not fatal
			return false, nil
		}
		if sa.Status.S3CredentialsSecretName == nil {
			return false, nil
		}

		// Fetch the secret
		secret, err := s.client.BucketServiceAccountSecrets().Get(ctx, *sa.Status.S3CredentialsSecretName)
		if err != nil {
			// Transient errors during polling are not fatal
			return false, nil
		}

		if secret.Data.AccessKeyID != nil && secret.Data.SecretAccessKey != nil {
			result = &S3Credentials{
				AccessKeyID:     *secret.Data.AccessKeyID,
				SecretAccessKey: *secret.Data.SecretAccessKey,
			}
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
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
		if errors.Is(err, rest.ErrNotFound) {
			// Resource deleted successfully
			return true, nil
		}
		// Still exists, keep polling
		return false, nil
	})
}
