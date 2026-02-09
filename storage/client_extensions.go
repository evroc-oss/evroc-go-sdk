// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package storage

// BucketServiceAccountSecrets returns the BucketServiceAccountSecrets service
func (c *Client) BucketServiceAccountSecrets() *BucketServiceAccountSecretsService {
	return &BucketServiceAccountSecretsService{client: c}
}
