// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package storage provides builder patterns for storage resources.
package storage

import (
	"context"

	storage "github.com/evroc-oss/evroc-go-sdk/types/storage"
)

// ============================================================================
// Bucket Builder
// ============================================================================

// BucketBuilder provides a fluent interface for creating Bucket resources.
type BucketBuilder struct {
	name                string
	objectRetentionMode string
	lockingDurationDays int32
	lockingMode         string
	labels              map[string]string
}

// NewBucketBuilder creates a new builder for Bucket.
func NewBucketBuilder(name string) *BucketBuilder {
	return &BucketBuilder{
		name: name,
	}
}

// WithObjectRetentionMode sets the object retention mode
// Options: "Disabled" (default), "Suspended", "Versioned", "Locking".
func (b *BucketBuilder) WithObjectRetentionMode(mode string) *BucketBuilder {
	b.objectRetentionMode = mode
	return b
}

// WithDefaultObjectLocking sets default object locking (only works when retention mode is "Locking")
// mode: "GOVERNANCE" or "COMPLIANCE"
// durationDays: number of days to lock objects by default.
func (b *BucketBuilder) WithDefaultObjectLocking(mode string, durationDays int32) *BucketBuilder {
	b.lockingMode = mode
	b.lockingDurationDays = durationDays
	return b
}

// WithLabels sets user-defined labels for the bucket.
func (b *BucketBuilder) WithLabels(labels map[string]string) *BucketBuilder {
	b.labels = labels
	return b
}

// Build creates the BucketRequest structure.
func (b *BucketBuilder) Build() *storage.BucketRequest {
	req := &storage.BucketRequest{
		ApiVersion: "storage/v1",
		Kind:       "Bucket",
		Metadata: storage.RegionalMetadataRequest{
			Name: &b.name,
		},
		Spec: storage.BucketSpec{},
	}

	if b.objectRetentionMode != "" {
		mode := storage.BucketSpecObjectRetentionMode(b.objectRetentionMode)
		req.Spec.ObjectRetentionMode = &mode
	}

	if b.lockingMode != "" && b.lockingDurationDays > 0 {
		req.Spec.DefaultObjectLocking = &storage.BucketSpecDefaultObjectLocking{
			DurationDays: b.lockingDurationDays,
			Mode:         storage.BucketSpecDefaultObjectLockingMode(b.lockingMode),
		}
	}

	// Add labels if specified
	if len(b.labels) > 0 {
		userLabels := storage.UserLabels(b.labels)
		req.Metadata.UserLabels = &userLabels
	}

	return req
}

// Create is a convenience method that builds and creates the bucket in one call.
func (b *BucketBuilder) Create(ctx context.Context, client *BucketsService) (*storage.Bucket, error) {
	req := b.Build()
	return client.Create(ctx, req)
}

// ============================================================================
// BucketServiceAccount Builder
// ============================================================================

// BucketServiceAccountBuilder provides a fluent interface for creating BucketServiceAccount resources.
type BucketServiceAccountBuilder struct {
	name    string
	buckets []string
	labels  map[string]string
}

// NewBucketServiceAccountBuilder creates a new builder for BucketServiceAccount.
func NewBucketServiceAccountBuilder(name string) *BucketServiceAccountBuilder {
	return &BucketServiceAccountBuilder{
		name:    name,
		buckets: []string{},
	}
}

// WithBucket adds a bucket that this service account can access.
func (b *BucketServiceAccountBuilder) WithBucket(bucketName string) *BucketServiceAccountBuilder {
	b.buckets = append(b.buckets, bucketName)
	return b
}

// WithBuckets adds multiple buckets that this service account can access.
func (b *BucketServiceAccountBuilder) WithBuckets(bucketNames ...string) *BucketServiceAccountBuilder {
	b.buckets = append(b.buckets, bucketNames...)
	return b
}

// WithLabels sets user-defined labels for the bucket service account.
func (b *BucketServiceAccountBuilder) WithLabels(labels map[string]string) *BucketServiceAccountBuilder {
	b.labels = labels
	return b
}

// Build creates the BucketServiceAccountRequest structure.
func (b *BucketServiceAccountBuilder) Build() *storage.BucketServiceAccountRequest {
	req := &storage.BucketServiceAccountRequest{
		ApiVersion: "storage/v1",
		Kind:       "BucketServiceAccount",
		Metadata: storage.RegionalMetadataRequest{
			Name: &b.name,
		},
		Spec: storage.BucketServiceAccountSpec{},
	}

	if len(b.buckets) > 0 {
		req.Spec.Buckets = &b.buckets
	}

	// Add labels if specified
	if len(b.labels) > 0 {
		userLabels := storage.UserLabels(b.labels)
		req.Metadata.UserLabels = &userLabels
	}

	return req
}

// Create is a convenience method that builds and creates the service account in one call.
func (b *BucketServiceAccountBuilder) Create(ctx context.Context, client *BucketServiceAccountsService) (*storage.BucketServiceAccount, error) {
	req := b.Build()
	return client.Create(ctx, req)
}
