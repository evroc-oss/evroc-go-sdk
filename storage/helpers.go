// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package storage

import (
	storagetypes "github.com/evroc-oss/evroc-go-sdk/types/storage"
)

// ============================================================================
// Status Helpers
// ============================================================================

// IsBucketReady returns true if the Bucket is in Ready condition.
func IsBucketReady(bucket *storagetypes.Bucket) bool {
	if bucket == nil || bucket.Status.Conditions == nil {
		return false
	}

	for _, cond := range *bucket.Status.Conditions {
		if cond.Type == "Ready" && string(cond.Status) == "True" {
			return true
		}
	}
	return false
}
