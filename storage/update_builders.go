// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"fmt"

	storage "github.com/evroc-oss/evroc-go-sdk/types/storage"
)

// Update Builders
//
// Update builders provide a fluent interface for modifying existing storage resources.
//
// Example - Updating bucket retention mode:
//
//	updated, err := UpdateBucket("my-bucket", bucketService).
//		SetRetentionMode(storage.Versioned).
//		Apply(ctx)
//
// Example - Adding buckets to a service account:
//
//	updated, err := UpdateBucketServiceAccount("my-sa", saService).
//		AddBucket("new-bucket").
//		Apply(ctx)

// BucketUpdateBuilder provides a fluent interface for updating Bucket resources.
//
// This builder simplifies updating bucket configuration such as retention modes
// and object locking policies.
type BucketUpdateBuilder struct {
	name          string
	service       *BucketsService
	retentionMode *storage.RetentionMode
	locking       *storage.BucketSpecDefaultObjectLocking
}

// NewBucketUpdateBuilder creates a new builder for updating a bucket.
func NewBucketUpdateBuilder(name string, service *BucketsService) *BucketUpdateBuilder {
	return &BucketUpdateBuilder{
		name:    name,
		service: service,
	}
}

// SetRetentionMode sets the object retention mode.
func (b *BucketUpdateBuilder) SetRetentionMode(mode storage.RetentionMode) *BucketUpdateBuilder {
	b.retentionMode = &mode
	return b
}

// SetObjectLocking sets the default object locking configuration.
func (b *BucketUpdateBuilder) SetObjectLocking(mode storage.LockingMode, durationDays int32) *BucketUpdateBuilder {
	b.locking = &storage.BucketSpecDefaultObjectLocking{
		Mode:         mode,
		DurationDays: durationDays,
	}
	return b
}

// Apply applies all pending updates to the bucket.
func (b *BucketUpdateBuilder) Apply(ctx context.Context) (*storage.Bucket, error) {
	if b.retentionMode == nil && b.locking == nil {
		return nil, fmt.Errorf("no updates to apply")
	}

	// Fetch current bucket state
	bucket, err := b.service.Get(ctx, b.name)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bucket: %w", err)
	}

	// Apply retention mode change
	if b.retentionMode != nil {
		bucket.Spec.ObjectRetentionMode = b.retentionMode
	}

	// Apply locking change
	if b.locking != nil {
		bucket.Spec.DefaultObjectLocking = b.locking
	}

	// Send the update
	updated, err := b.service.Patch(ctx, b.name, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to update bucket: %w", err)
	}

	return updated, nil
}

// BucketServiceAccountUpdateBuilder provides a fluent interface for updating BucketServiceAccount resources.
type BucketServiceAccountUpdateBuilder struct {
	name    string
	service *BucketServiceAccountsService
	buckets *[]string
}

// NewBucketServiceAccountUpdateBuilder creates a new builder for updating a bucket service account.
func NewBucketServiceAccountUpdateBuilder(name string, service *BucketServiceAccountsService) *BucketServiceAccountUpdateBuilder {
	return &BucketServiceAccountUpdateBuilder{
		name:    name,
		service: service,
	}
}

// SetBuckets replaces the list of accessible buckets.
func (b *BucketServiceAccountUpdateBuilder) SetBuckets(buckets []string) *BucketServiceAccountUpdateBuilder {
	b.buckets = &buckets
	return b
}

// AddBucket adds a bucket to the access list.
func (b *BucketServiceAccountUpdateBuilder) AddBucket(bucket string) *BucketServiceAccountUpdateBuilder {
	if b.buckets == nil {
		b.buckets = &[]string{}
	}
	*b.buckets = append(*b.buckets, bucket)
	return b
}

// Apply applies all pending updates to the bucket service account.
func (b *BucketServiceAccountUpdateBuilder) Apply(ctx context.Context) (*storage.BucketServiceAccount, error) {
	if b.buckets == nil {
		return nil, fmt.Errorf("no updates to apply")
	}

	// Fetch current service account state
	sa, err := b.service.Get(ctx, b.name)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bucket service account: %w", err)
	}

	// Apply bucket list change
	sa.Spec.Buckets = b.buckets

	// Send the update
	updated, err := b.service.Patch(ctx, b.name, sa)
	if err != nil {
		return nil, fmt.Errorf("failed to update bucket service account: %w", err)
	}

	return updated, nil
}

// UpdateBucket creates an update builder for a bucket.
// This is a convenience function for the common case.
func UpdateBucket(name string, service *BucketsService) *BucketUpdateBuilder {
	return NewBucketUpdateBuilder(name, service)
}

// UpdateBucketServiceAccount creates an update builder for a bucket service account.
// This is a convenience function for the common case.
func UpdateBucketServiceAccount(name string, service *BucketServiceAccountsService) *BucketServiceAccountUpdateBuilder {
	return NewBucketServiceAccountUpdateBuilder(name, service)
}
