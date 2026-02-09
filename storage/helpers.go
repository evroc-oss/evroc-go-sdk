// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package storage

import (
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
