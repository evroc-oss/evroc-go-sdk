// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"fmt"

	"github.com/evroc-oss/evroc-go-sdk/metrics"
	storagetypes "github.com/evroc-oss/evroc-go-sdk/types/storage"
)

// ============================================================================
// Metrics Support
// ============================================================================

// WithMetrics enables metrics collection for this storage client.
// Returns the client to allow chaining.
func (c *Client) WithMetrics(m *metrics.Manager) *Client {
	c.metrics = m
	return c
}

// ============================================================================
// Status Helpers
// ============================================================================

// IsBucketReady returns true if the Bucket is in Ready condition.
func IsBucketReady(bucket *storagetypes.Bucket) bool {
	if bucket == nil || bucket.Status.Conditions == nil {
		return false
	}

	for _, cond := range *bucket.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}
	return false
}

// IsBucketServiceAccountReady returns true if the BucketServiceAccount is in Ready condition.
func IsBucketServiceAccountReady(bsa *storagetypes.BucketServiceAccount) bool {
	if bsa == nil || bsa.Status.Conditions == nil {
		return false
	}

	for _, cond := range *bsa.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}
	return false
}

// IsFileStoreAvailable returns true if the FileStore status is Available.
func IsFileStoreAvailable(fs *storagetypes.Filestore) bool {
	if fs == nil || fs.Status.Status == nil {
		return false
	}
	return *fs.Status.Status == storagetypes.Available
}

// ============================================================================
// S3 Endpoint Helper
// ============================================================================

// GetS3Endpoint returns the S3 endpoint for the client's configured region.
func (c *Client) GetS3Endpoint() string {
	return fmt.Sprintf("s3.%s.evroc.com", c.parent.DefaultRegion())
}
