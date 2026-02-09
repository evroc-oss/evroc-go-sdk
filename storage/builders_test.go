// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package storage

import "testing"

func TestBucketBuilder(t *testing.T) {
	req := NewBucketBuilder("test-bucket").
		WithObjectRetentionMode("Locking").
		WithDefaultObjectLocking("GOVERNANCE", 30).
		WithLabels(map[string]string{"env": "prod"}).
		Build()

	// Validate basic fields
	if req.Kind != "Bucket" {
		t.Errorf("Expected Kind 'Bucket', got %s", req.Kind)
	}
	if *req.Metadata.Id != "test-bucket" {
		t.Errorf("Expected Id 'test-bucket', got %s", *req.Metadata.Id)
	}

	// Validate object retention mode
	if req.Spec.ObjectRetentionMode == nil {
		t.Error("ObjectRetentionMode should not be nil")
	} else if *req.Spec.ObjectRetentionMode != "Locking" {
		t.Errorf("Expected retention mode 'Locking', got %s", *req.Spec.ObjectRetentionMode)
	}

	// Validate default object locking
	if req.Spec.DefaultObjectLocking == nil {
		t.Error("DefaultObjectLocking should not be nil")
	} else {
		if req.Spec.DefaultObjectLocking.Mode != "GOVERNANCE" {
			t.Errorf("Expected locking mode 'GOVERNANCE', got %s", req.Spec.DefaultObjectLocking.Mode)
		}
		if req.Spec.DefaultObjectLocking.DurationDays != 30 {
			t.Errorf("Expected duration 30 days, got %d", req.Spec.DefaultObjectLocking.DurationDays)
		}
	}

	// Validate labels
	if req.Metadata.UserLabels == nil {
		t.Error("UserLabels should not be nil")
	} else if (*req.Metadata.UserLabels)["env"] != "prod" {
		t.Errorf("Expected label env='prod', got %s", (*req.Metadata.UserLabels)["env"])
	}
}

func TestBucketServiceAccountBuilder(t *testing.T) {
	req := NewBucketServiceAccountBuilder("test-sa").
		WithBucket("bucket1").
		WithBuckets("bucket2", "bucket3").
		WithLabels(map[string]string{"env": "prod", "team": "platform"}).
		Build()

	// Validate basic fields
	if req.Kind != "BucketServiceAccount" {
		t.Errorf("Expected Kind 'BucketServiceAccount', got %s", req.Kind)
	}
	if *req.Metadata.Id != "test-sa" {
		t.Errorf("Expected Id 'test-sa', got %s", *req.Metadata.Id)
	}

	// Validate buckets
	if req.Spec.Buckets == nil {
		t.Error("Buckets should not be nil")
	} else {
		buckets := *req.Spec.Buckets
		if len(buckets) != 3 {
			t.Errorf("Expected 3 buckets, got %d", len(buckets))
		} else {
			if buckets[0] != "bucket1" {
				t.Errorf("Expected bucket 'bucket1', got %s", buckets[0])
			}
			if buckets[1] != "bucket2" {
				t.Errorf("Expected bucket 'bucket2', got %s", buckets[1])
			}
			if buckets[2] != "bucket3" {
				t.Errorf("Expected bucket 'bucket3', got %s", buckets[2])
			}
		}
	}

	// Validate labels
	if req.Metadata.UserLabels == nil {
		t.Error("UserLabels should not be nil")
	} else {
		if (*req.Metadata.UserLabels)["env"] != "prod" {
			t.Errorf("Expected label env='prod', got %s", (*req.Metadata.UserLabels)["env"])
		}
		if (*req.Metadata.UserLabels)["team"] != "platform" {
			t.Errorf("Expected label team='platform', got %s", (*req.Metadata.UserLabels)["team"])
		}
	}
}
