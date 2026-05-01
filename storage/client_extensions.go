// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package storage

// BucketServiceAccountSecrets returns the BucketServiceAccountSecrets service
func (c *Client) BucketServiceAccountSecrets() *BucketServiceAccountSecretsService {
	return &BucketServiceAccountSecretsService{client: c}
}
