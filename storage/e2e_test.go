// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/internal/e2etest"
	"github.com/evroc-oss/evroc-go-sdk/storage"
	storagetypes "github.com/evroc-oss/evroc-go-sdk/types/storage"
)

func TestE2E_Bucket_Lifecycle(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client := e2etest.NewClient(t)
	bucketName := e2etest.RandomName("bucket")

	t.Logf("Creating bucket: %s", bucketName)

	// Create bucket
	bucket, err := storage.NewBucketBuilder(bucketName).
		Create(ctx, client.Storage().Buckets())

	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	bucketID := e2etest.MustGetID(t, bucket.Metadata.Id, "bucket")
	t.Logf("Created bucket with ID: %s", bucketID)

	bucketDeleted := false
	e2etest.DeferCleanup(t, ctx, client.Storage().Buckets().Delete, bucketID, "bucket", &bucketDeleted)

	// Wait for bucket to be ready
	t.Logf("Waiting for bucket to be ready...")
	if _, err := client.Storage().Buckets().WaitForReady(ctx, bucketID, 2*time.Minute); err != nil {
		t.Fatalf("bucket never became ready: %v", err)
	}

	// Read bucket
	t.Logf("Reading bucket: %s", bucketID)
	retrieved, err := client.Storage().Buckets().Get(ctx, bucketID)
	if err != nil {
		t.Fatalf("failed to get bucket: %v", err)
	}

	if retrieved.Metadata.Id != bucketID {
		t.Errorf("expected bucket ID %s, got %v", bucketID, retrieved.Metadata.Id)
	}

	// List buckets - should include our bucket
	t.Logf("Listing buckets")
	buckets, err := client.Storage().Buckets().List(ctx)
	if err != nil {
		t.Fatalf("failed to list buckets: %v", err)
	}
	e2etest.AssertInList(t, buckets.Items, bucketID, func(b storagetypes.Bucket) string { return b.Metadata.Id }, "bucket")

	// Delete bucket
	t.Logf("Deleting bucket: %s", bucketID)
	if err := client.Storage().Buckets().Delete(ctx, bucketID); err != nil {
		t.Fatalf("failed to delete bucket: %v", err)
	}
	bucketDeleted = true

	// Verify bucket was deleted
	t.Logf("Verifying bucket deletion")
	e2etest.AssertDeleted(t, ctx, func(ctx context.Context, id string) (any, error) {
		return client.Storage().Buckets().Get(ctx, id)
	}, bucketID, "bucket")

	t.Logf("Bucket lifecycle test completed successfully")
}

func TestE2E_BucketServiceAccount_Lifecycle(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client := e2etest.NewClient(t)
	bucketName := e2etest.RandomName("bucket")
	saName := e2etest.RandomName("sa")

	t.Logf("Creating bucket: %s", bucketName)

	// Create bucket first
	bucket, err := storage.NewBucketBuilder(bucketName).
		Create(ctx, client.Storage().Buckets())

	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	bucketID := e2etest.MustGetID(t, bucket.Metadata.Id, "bucket")
	t.Logf("Created bucket with ID: %s", bucketID)

	bucketDeleted := false
	e2etest.DeferCleanup(t, ctx, client.Storage().Buckets().Delete, bucketID, "bucket", &bucketDeleted)

	// Wait for bucket to be ready before creating service account
	t.Logf("Waiting for bucket to be ready...")
	if _, err := client.Storage().Buckets().WaitForReady(ctx, bucketID, 2*time.Minute); err != nil {
		t.Fatalf("bucket never became ready: %v", err)
	}

	t.Logf("Creating service account: %s", saName)

	// Create service account
	sa, err := storage.NewBucketServiceAccountBuilder(saName).
		WithBucket(bucketID).
		Create(ctx, client.Storage().BucketServiceAccounts())

	if err != nil {
		t.Fatalf("failed to create service account: %v", err)
	}

	saID := e2etest.MustGetID(t, sa.Metadata.Id, "service account")
	t.Logf("Created service account with ID: %s", saID)

	saDeleted := false
	e2etest.DeferCleanup(t, ctx, client.Storage().BucketServiceAccounts().Delete, saID, "service account", &saDeleted)

	// Wait for service account to be ready
	t.Logf("Waiting for service account to be ready...")
	if _, err := client.Storage().BucketServiceAccounts().WaitForReady(ctx, saID, 2*time.Minute); err != nil {
		t.Fatalf("service account never became ready: %v", err)
	}

	// Verify service account was created with correct bucket reference
	if sa.Spec.Buckets == nil || len(*sa.Spec.Buckets) != 1 {
		t.Error("expected 1 bucket in service account")
	} else if (*sa.Spec.Buckets)[0] != bucketID {
		t.Errorf("expected bucket reference %s, got %s", bucketID, (*sa.Spec.Buckets)[0])
	}

	// Read service account
	t.Logf("Reading service account: %s", saID)
	retrieved, err := client.Storage().BucketServiceAccounts().Get(ctx, saID)
	if err != nil {
		t.Fatalf("failed to get service account: %v", err)
	}

	if retrieved.Metadata.Id != saID {
		t.Errorf("expected service account ID %s, got %v", saID, retrieved.Metadata.Id)
	}

	// List service accounts - should include ours
	t.Logf("Listing service accounts")
	sas, err := client.Storage().BucketServiceAccounts().List(ctx)
	if err != nil {
		t.Fatalf("failed to list service accounts: %v", err)
	}
	e2etest.AssertInList(t, sas.Items, saID, func(s storagetypes.BucketServiceAccount) string { return s.Metadata.Id }, "service account")

	// Delete service account
	t.Logf("Deleting service account: %s", saID)
	if err := client.Storage().BucketServiceAccounts().Delete(ctx, saID); err != nil {
		t.Fatalf("failed to delete service account: %v", err)
	}
	saDeleted = true

	t.Logf("Service account lifecycle test completed successfully")
}
