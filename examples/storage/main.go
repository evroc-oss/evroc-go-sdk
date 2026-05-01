// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package main demonstrates comprehensive storage API usage.
// Covers: Buckets, BucketServiceAccounts
//
// NOTE: This example demonstrates the evroc SDK APIs only. Buckets must be empty
// before they can be deleted. To delete bucket contents, use an S3-compatible client
// (see examples/vm-backup-to-storage for a complete example with MinIO SDK, and
// examples/storage-public-urls for public URL generation examples).
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/storage"
	storagetypes "github.com/evroc-oss/evroc-go-sdk/types/storage"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== Comprehensive Storage API Examples ===")
	fmt.Println()

	// Run all examples
	if err := runBucketExamples(ctx, client); err != nil {
		log.Printf("Bucket examples failed: %v", err)
	}

	if err := runBucketServiceAccountExamples(ctx, client); err != nil {
		log.Printf("Bucket service account examples failed: %v", err)
	}

	fmt.Println("\n=== All Storage Examples Complete ===")
}

// runBucketExamples demonstrates all bucket operations.
func runBucketExamples(ctx context.Context, client *evroc.Client) error {
	fmt.Println("--- Bucket Examples ---")

	// Example 1: Create a simple bucket
	fmt.Println("\n1. Creating a simple bucket...")
	simpleBucket := storage.NewBucketBuilder("sdk-bucket-simple").Build()

	createdSimple, err := client.Storage().Buckets().Create(ctx, simpleBucket)
	if err != nil {
		return fmt.Errorf("failed to create simple bucket: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", createdSimple.Metadata.Id)

	// Example 2: Create a bucket with versioning suspended
	fmt.Println("\n2. Creating bucket with versioning suspended...")
	versionedBucket := storage.NewBucketBuilder("sdk-bucket-versioned").
		WithObjectRetentionMode(storage.Suspended).
		Build()

	createdVersioned, err := client.Storage().Buckets().Create(ctx, versionedBucket)
	if err != nil {
		return fmt.Errorf("failed to create versioned bucket: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (retention mode: %s)\n",
		createdVersioned.Metadata.Id,
		*createdVersioned.Spec.ObjectRetentionMode)

	// Example 3: Create a bucket with object locking
	fmt.Println("\n3. Creating bucket with object locking...")
	lockingBucket := storage.NewBucketBuilder("sdk-bucket-locking").
		WithObjectRetentionMode(storage.Locking).
		WithDefaultObjectLocking(storage.Soft, 30).
		Build()

	createdLocking, err := client.Storage().Buckets().Create(ctx, lockingBucket)
	if err != nil {
		return fmt.Errorf("failed to create locking bucket: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (locking mode: %s, duration: %d days)\n",
		createdLocking.Metadata.Id,
		createdLocking.Spec.DefaultObjectLocking.Mode,
		createdLocking.Spec.DefaultObjectLocking.DurationDays)

	// Example 4: Create a bucket with compliance locking
	fmt.Println("\n4. Creating bucket with compliance locking...")
	complianceBucket := storage.NewBucketBuilder("sdk-bucket-compliance").
		WithObjectRetentionMode(storage.Locking).
		WithDefaultObjectLocking(storage.Immutable, 90).
		Build()

	createdCompliance, err := client.Storage().Buckets().Create(ctx, complianceBucket)
	if err != nil {
		return fmt.Errorf("failed to create compliance bucket: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (compliance locking: %d days)\n",
		createdCompliance.Metadata.Id,
		createdCompliance.Spec.DefaultObjectLocking.DurationDays)

	// Example 5: List all buckets
	fmt.Println("\n5. Listing all buckets...")
	buckets, err := client.Storage().Buckets().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list buckets: %w", err)
	}
	fmt.Printf("   Found %d buckets:\n", len(buckets.Items))
	for _, bucket := range buckets.Items {
		retentionMode := "Disabled"
		if bucket.Spec.ObjectRetentionMode != nil {
			retentionMode = string(*bucket.Spec.ObjectRetentionMode)
		}
		fmt.Printf("   - %s (retention: %s)\n", bucket.Metadata.Id, retentionMode)
	}

	// Example 6: Get a specific bucket
	fmt.Println("\n6. Getting specific bucket...")
	bucket, err := client.Storage().Buckets().Get(ctx, "sdk-bucket-locking")
	if err != nil {
		return fmt.Errorf("failed to get bucket: %w", err)
	}
	fmt.Printf("   ✓ Bucket: %s\n", bucket.Metadata.Id)
	if bucket.Spec.ObjectRetentionMode != nil {
		fmt.Printf("     Retention Mode: %s\n", *bucket.Spec.ObjectRetentionMode)
	}
	if bucket.Spec.DefaultObjectLocking != nil {
		fmt.Printf("     Locking Mode: %s\n", bucket.Spec.DefaultObjectLocking.Mode)
		fmt.Printf("     Locking Duration: %d days\n", bucket.Spec.DefaultObjectLocking.DurationDays)
	}

	// Example 7: Update bucket retention mode
	fmt.Println("\n7. Updating bucket object retention mode...")
	bucketToUpdate, err := client.Storage().Buckets().Get(ctx, "sdk-bucket-simple")
	if err != nil {
		log.Printf("   Warning: Failed to get bucket for update: %v", err)
	} else {
		newMode := storagetypes.Versioned
		bucketToUpdate.Spec.ObjectRetentionMode = &newMode
		updatedBucket, err := client.Storage().Buckets().Patch(ctx, "sdk-bucket-simple", bucketToUpdate)
		if err != nil {
			log.Printf("   Warning: Update failed (may not be supported): %v", err)
		} else {
			fmt.Printf("   ✓ Updated: %s (new retention mode: %s)\n",
				updatedBucket.Metadata.Id,
				*updatedBucket.Spec.ObjectRetentionMode)
		}
	}

	// Example 8: Delete a bucket (must be empty first)
	fmt.Println("\n8. Deleting a bucket...")
	fmt.Println("   Note: In production, you must delete all objects from the bucket first")
	fmt.Println("   before deleting the bucket. This example assumes the bucket is empty.")
	err = client.Storage().Buckets().Delete(ctx, "sdk-bucket-compliance")
	if err != nil {
		log.Printf("   Warning: Failed to delete bucket (may contain objects): %v", err)
		log.Println("   To delete a bucket with objects, use an S3 client to delete all objects first")
	} else {
		fmt.Println("   ✓ Deleted sdk-bucket-compliance")
	}

	return nil
}

// runBucketServiceAccountExamples demonstrates all bucket service account operations.
func runBucketServiceAccountExamples(ctx context.Context, client *evroc.Client) error {
	fmt.Println("\n--- Bucket Service Account Examples ---")

	// Example 1: Create a service account for a single bucket
	fmt.Println("\n1. Creating service account for single bucket...")
	saSingle := storage.NewBucketServiceAccountBuilder("sdk-sa-single").
		WithBucket("sdk-bucket-simple").
		Build()

	createdSASingle, err := client.Storage().BucketServiceAccounts().Create(ctx, saSingle)
	if err != nil {
		return fmt.Errorf("failed to create single bucket SA: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", createdSASingle.Metadata.Id)

	// Example 2: Create a service account for multiple buckets
	fmt.Println("\n2. Creating service account for multiple buckets...")
	saMulti := storage.NewBucketServiceAccountBuilder("sdk-sa-multi").
		WithBuckets("sdk-bucket-simple", "sdk-bucket-versioned", "sdk-bucket-locking").
		Build()

	createdSAMulti, err := client.Storage().BucketServiceAccounts().Create(ctx, saMulti)
	if err != nil {
		return fmt.Errorf("failed to create multi bucket SA: %w", err)
	}
	bucketCount := 0
	if createdSAMulti.Spec.Buckets != nil {
		bucketCount = len(*createdSAMulti.Spec.Buckets)
	}
	fmt.Printf("   ✓ Created: %s (buckets: %d)\n", createdSAMulti.Metadata.Id, bucketCount)

	// Example 3: Create a service account using WithBucket method multiple times
	fmt.Println("\n3. Creating service account by adding buckets individually...")
	saIndividual := storage.NewBucketServiceAccountBuilder("sdk-sa-individual").
		WithBucket("sdk-bucket-simple").
		WithBucket("sdk-bucket-versioned").
		Build()

	createdSAIndividual, err := client.Storage().BucketServiceAccounts().Create(ctx, saIndividual)
	if err != nil {
		return fmt.Errorf("failed to create individual bucket SA: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", createdSAIndividual.Metadata.Id)

	// Example 4: List all bucket service accounts
	fmt.Println("\n4. Listing all bucket service accounts...")
	serviceAccounts, err := client.Storage().BucketServiceAccounts().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list service accounts: %w", err)
	}
	fmt.Printf("   Found %d service accounts:\n", len(serviceAccounts.Items))
	for _, sa := range serviceAccounts.Items {
		bucketCount := 0
		if sa.Spec.Buckets != nil {
			bucketCount = len(*sa.Spec.Buckets)
		}
		fmt.Printf("   - %s (buckets: %d)\n", sa.Metadata.Id, bucketCount)
	}

	// Example 5: Get a specific service account
	fmt.Println("\n5. Getting specific service account...")
	sa, err := client.Storage().BucketServiceAccounts().Get(ctx, "sdk-sa-multi")
	if err != nil {
		return fmt.Errorf("failed to get service account: %w", err)
	}
	fmt.Printf("   ✓ Service Account: %s\n", sa.Metadata.Id)
	if sa.Spec.Buckets != nil {
		fmt.Printf("     Buckets:\n")
		for _, bucket := range *sa.Spec.Buckets {
			fmt.Printf("       - %s\n", bucket)
		}
	}

	// Example 6: Update service account (add more buckets)
	fmt.Println("\n6. Updating service account to add more buckets...")
	saToUpdate, err := client.Storage().BucketServiceAccounts().Get(ctx, "sdk-sa-single")
	if err != nil {
		log.Printf("   Warning: Failed to get service account for update: %v", err)
	} else {
		newBuckets := []string{"sdk-bucket-simple", "sdk-bucket-locking"}
		saToUpdate.Spec.Buckets = &newBuckets
		updatedSA, err := client.Storage().BucketServiceAccounts().Patch(ctx, "sdk-sa-single", saToUpdate)
		if err != nil {
			log.Printf("   Warning: Update failed (may not be supported): %v", err)
		} else {
			bucketCount := 0
			if updatedSA.Spec.Buckets != nil {
				bucketCount = len(*updatedSA.Spec.Buckets)
			}
			fmt.Printf("   ✓ Updated: %s (now has %d buckets)\n", updatedSA.Metadata.Id, bucketCount)
		}
	}

	// Example 7: Delete a service account
	fmt.Println("\n7. Deleting a service account...")
	err = client.Storage().BucketServiceAccounts().Delete(ctx, "sdk-sa-individual")
	if err != nil {
		return fmt.Errorf("failed to delete service account: %w", err)
	}
	fmt.Println("   ✓ Deleted sdk-sa-individual")

	return nil
}
