// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package storage

import (
	"testing"
	"time"

	storagetypes "github.com/evroc-oss/evroc-go-sdk/types/storage"
)

func TestIsBucketReady(t *testing.T) {
	t.Run("nil bucket", func(t *testing.T) {
		if IsBucketReady(nil) {
			t.Error("nil bucket should not be ready")
		}
	})

	t.Run("bucket with no conditions", func(t *testing.T) {
		bucket := &storagetypes.Bucket{}
		if IsBucketReady(bucket) {
			t.Error("bucket with no conditions should not be ready")
		}
	})

	t.Run("bucket with Ready condition True", func(t *testing.T) {
		conditions := []storagetypes.BucketStatusConditionsItem{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: time.Now(),
				Reason:             "BucketReady",
				Message:            "Bucket is ready",
			},
		}
		bucket := &storagetypes.Bucket{
			Status: storagetypes.BucketStatus{
				Conditions: &conditions,
			},
		}
		if !IsBucketReady(bucket) {
			t.Error("bucket with Ready=True should be ready")
		}
	})

	t.Run("bucket with Ready condition False", func(t *testing.T) {
		conditions := []storagetypes.BucketStatusConditionsItem{
			{
				Type:               "Ready",
				Status:             "False",
				LastTransitionTime: time.Now(),
				Reason:             "Creating",
				Message:            "Bucket is being created",
			},
		}
		bucket := &storagetypes.Bucket{
			Status: storagetypes.BucketStatus{
				Conditions: &conditions,
			},
		}
		if IsBucketReady(bucket) {
			t.Error("bucket with Ready=False should not be ready")
		}
	})

	t.Run("bucket with multiple conditions", func(t *testing.T) {
		conditions := []storagetypes.BucketStatusConditionsItem{
			{
				Type:               "Allocated",
				Status:             "True",
				LastTransitionTime: time.Now(),
				Reason:             "Allocated",
				Message:            "Storage allocated",
			},
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: time.Now(),
				Reason:             "Ready",
				Message:            "Bucket ready",
			},
		}
		bucket := &storagetypes.Bucket{
			Status: storagetypes.BucketStatus{
				Conditions: &conditions,
			},
		}
		if !IsBucketReady(bucket) {
			t.Error("bucket with Ready=True should be ready even with multiple conditions")
		}
	})

	t.Run("bucket with Ready condition Unknown", func(t *testing.T) {
		conditions := []storagetypes.BucketStatusConditionsItem{
			{
				Type:               "Ready",
				Status:             "Unknown",
				LastTransitionTime: time.Now(),
				Reason:             "Unknown",
				Message:            "Status unknown",
			},
		}
		bucket := &storagetypes.Bucket{
			Status: storagetypes.BucketStatus{
				Conditions: &conditions,
			},
		}
		if IsBucketReady(bucket) {
			t.Error("bucket with Ready=Unknown should not be ready")
		}
	})
}
