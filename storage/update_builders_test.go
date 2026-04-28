// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"testing"

	storage "github.com/evroc-oss/evroc-go-sdk/types/storage"
)

func TestBucketUpdateBuilder_BuilderMethods(t *testing.T) {
	builder := &BucketUpdateBuilder{
		name: "test-bucket",
	}

	// Test SetRetentionMode
	builder.SetRetentionMode(storage.Versioned)
	if builder.retentionMode == nil || *builder.retentionMode != storage.Versioned {
		t.Error("SetRetentionMode() should set retentionMode")
	}

	// Test SetObjectLocking
	builder.SetObjectLocking(storage.Immutable, 30)
	if builder.locking == nil {
		t.Fatal("SetObjectLocking() should set locking")
	}
	if builder.locking.Mode != storage.Immutable {
		t.Error("SetObjectLocking() should set mode to Immutable")
	}
	if builder.locking.DurationDays != 30 {
		t.Error("SetObjectLocking() should set duration to 30")
	}
}

func TestBucketServiceAccountUpdateBuilder_BuilderMethods(t *testing.T) {
	builder := &BucketServiceAccountUpdateBuilder{
		name: "test-sa",
	}

	// Test SetBuckets
	buckets := []string{"bucket1", "bucket2"}
	builder.SetBuckets(buckets)
	if builder.buckets == nil || len(*builder.buckets) != 2 {
		t.Error("SetBuckets() should set buckets")
	}

	// Test AddBucket
	builder2 := &BucketServiceAccountUpdateBuilder{
		name: "test-sa",
	}
	builder2.AddBucket("bucket3")
	if builder2.buckets == nil || len(*builder2.buckets) != 1 || (*builder2.buckets)[0] != "bucket3" {
		t.Error("AddBucket() should add bucket to list")
	}
}
