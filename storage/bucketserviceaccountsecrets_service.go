// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package storage

import (
	"context"
	"path"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/types/storage"
)

const (
	resourceBucketServiceAccountSecrets = "bucketServiceAccountSecrets"
)

// BucketServiceAccountSecretsService handles operations for bucket service account secrets
type BucketServiceAccountSecretsService struct {
	client *Client
}

// Get retrieves the secret credentials for a bucket service account
func (s *BucketServiceAccountSecretsService) Get(ctx context.Context, name string) (*storage.Bucketserviceaccountsecret, error) {
	resourcePath := path.Join(
		"/storage",
		apiVersion,
		"projects",
		s.client.parent.DefaultProject(),
		"regions",
		s.client.parent.DefaultRegion(),
		resourceBucketServiceAccountSecrets,
		name,
	)

	return rest.GetResource[*storage.Bucketserviceaccountsecret](ctx, s.client.rest, resourcePath)
}
