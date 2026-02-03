// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package main demonstrates comprehensive storage API usage.
// Covers: Buckets, BucketServiceAccounts
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
	simpleBucket := storage.NewBucketBuilder("example-bucket-simple").Build()

	createdSimple, err := client.Storage().Buckets().Create(ctx, simpleBucket)
	if err != nil {
		return fmt.Errorf("failed to create simple bucket: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", *createdSimple.Metadata.Name)

	// Example 2: Create a bucket with versioning suspended
	fmt.Println("\n2. Creating bucket with versioning suspended...")
	versionedBucket := storage.NewBucketBuilder("example-bucket-versioned").
		WithObjectRetentionMode("Suspended").
		Build()

	createdVersioned, err := client.Storage().Buckets().Create(ctx, versionedBucket)
	if err != nil {
		return fmt.Errorf("failed to create versioned bucket: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (retention mode: %s)\n",
		*createdVersioned.Metadata.Name,
		*createdVersioned.Spec.ObjectRetentionMode)

	// Example 3: Create a bucket with object locking
	fmt.Println("\n3. Creating bucket with object locking...")
	lockingBucket := storage.NewBucketBuilder("example-bucket-locking").
		WithObjectRetentionMode("Locking").
		WithDefaultObjectLocking("GOVERNANCE", 30).
		Build()

	createdLocking, err := client.Storage().Buckets().Create(ctx, lockingBucket)
	if err != nil {
		return fmt.Errorf("failed to create locking bucket: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (locking mode: %s, duration: %d days)\n",
		*createdLocking.Metadata.Name,
		createdLocking.Spec.DefaultObjectLocking.Mode,
		createdLocking.Spec.DefaultObjectLocking.DurationDays)

	// Example 4: Create a bucket with compliance locking
	fmt.Println("\n4. Creating bucket with compliance locking...")
	complianceBucket := storage.NewBucketBuilder("example-bucket-compliance").
		WithObjectRetentionMode("Locking").
		WithDefaultObjectLocking("COMPLIANCE", 90).
		Build()

	createdCompliance, err := client.Storage().Buckets().Create(ctx, complianceBucket)
	if err != nil {
		return fmt.Errorf("failed to create compliance bucket: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (compliance locking: %d days)\n",
		*createdCompliance.Metadata.Name,
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
		fmt.Printf("   - %s (retention: %s)\n", *bucket.Metadata.Name, retentionMode)
	}

	// Example 6: Get a specific bucket
	fmt.Println("\n6. Getting specific bucket...")
	bucket, err := client.Storage().Buckets().Get(ctx, "example-bucket-locking")
	if err != nil {
		return fmt.Errorf("failed to get bucket: %w", err)
	}
	fmt.Printf("   ✓ Bucket: %s\n", *bucket.Metadata.Name)
	if bucket.Spec.ObjectRetentionMode != nil {
		fmt.Printf("     Retention Mode: %s\n", *bucket.Spec.ObjectRetentionMode)
	}
	if bucket.Spec.DefaultObjectLocking != nil {
		fmt.Printf("     Locking Mode: %s\n", bucket.Spec.DefaultObjectLocking.Mode)
		fmt.Printf("     Locking Duration: %d days\n", bucket.Spec.DefaultObjectLocking.DurationDays)
	}

	// Example 7: Update bucket retention mode
	fmt.Println("\n7. Updating bucket object retention mode...")
	bucketToUpdate, err := client.Storage().Buckets().Get(ctx, "example-bucket-simple")
	if err != nil {
		log.Printf("   Warning: Failed to get bucket for update: %v", err)
	} else {
		newMode := storagetypes.Versioned
		bucketToUpdate.Spec.ObjectRetentionMode = &newMode
		updatedBucket, err := client.Storage().Buckets().Update(ctx, "example-bucket-simple", bucketToUpdate)
		if err != nil {
			log.Printf("   Warning: Update failed (may not be supported): %v", err)
		} else {
			fmt.Printf("   ✓ Updated: %s (new retention mode: %s)\n",
				*updatedBucket.Metadata.Name,
				*updatedBucket.Spec.ObjectRetentionMode)
		}
	}

	// Example 8: Delete a bucket
	fmt.Println("\n8. Deleting a bucket...")
	err = client.Storage().Buckets().Delete(ctx, "example-bucket-compliance")
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}
	fmt.Println("   ✓ Deleted example-bucket-compliance")

	return nil
}

// runBucketServiceAccountExamples demonstrates all bucket service account operations.
func runBucketServiceAccountExamples(ctx context.Context, client *evroc.Client) error {
	fmt.Println("\n--- Bucket Service Account Examples ---")

	// Example 1: Create a service account for a single bucket
	fmt.Println("\n1. Creating service account for single bucket...")
	saSingle := storage.NewBucketServiceAccountBuilder("example-sa-single").
		WithBucket("example-bucket-simple").
		Build()

	createdSASingle, err := client.Storage().BucketServiceAccounts().Create(ctx, saSingle)
	if err != nil {
		return fmt.Errorf("failed to create single bucket SA: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", *createdSASingle.Metadata.Name)

	// Example 2: Create a service account for multiple buckets
	fmt.Println("\n2. Creating service account for multiple buckets...")
	saMulti := storage.NewBucketServiceAccountBuilder("example-sa-multi").
		WithBuckets("example-bucket-simple", "example-bucket-versioned", "example-bucket-locking").
		Build()

	createdSAMulti, err := client.Storage().BucketServiceAccounts().Create(ctx, saMulti)
	if err != nil {
		return fmt.Errorf("failed to create multi bucket SA: %w", err)
	}
	bucketCount := 0
	if createdSAMulti.Spec.Buckets != nil {
		bucketCount = len(*createdSAMulti.Spec.Buckets)
	}
	fmt.Printf("   ✓ Created: %s (buckets: %d)\n", *createdSAMulti.Metadata.Name, bucketCount)

	// Example 3: Create a service account using WithBucket method multiple times
	fmt.Println("\n3. Creating service account by adding buckets individually...")
	saIndividual := storage.NewBucketServiceAccountBuilder("example-sa-individual").
		WithBucket("example-bucket-simple").
		WithBucket("example-bucket-versioned").
		Build()

	createdSAIndividual, err := client.Storage().BucketServiceAccounts().Create(ctx, saIndividual)
	if err != nil {
		return fmt.Errorf("failed to create individual bucket SA: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", *createdSAIndividual.Metadata.Name)

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
		fmt.Printf("   - %s (buckets: %d)\n", *sa.Metadata.Name, bucketCount)
	}

	// Example 5: Get a specific service account
	fmt.Println("\n5. Getting specific service account...")
	sa, err := client.Storage().BucketServiceAccounts().Get(ctx, "example-sa-multi")
	if err != nil {
		return fmt.Errorf("failed to get service account: %w", err)
	}
	fmt.Printf("   ✓ Service Account: %s\n", *sa.Metadata.Name)
	if sa.Spec.Buckets != nil {
		fmt.Printf("     Buckets:\n")
		for _, bucket := range *sa.Spec.Buckets {
			fmt.Printf("       - %s\n", bucket)
		}
	}

	// Example 6: Update service account (add more buckets)
	fmt.Println("\n6. Updating service account to add more buckets...")
	saToUpdate, err := client.Storage().BucketServiceAccounts().Get(ctx, "example-sa-single")
	if err != nil {
		log.Printf("   Warning: Failed to get service account for update: %v", err)
	} else {
		newBuckets := []string{"example-bucket-simple", "example-bucket-locking"}
		saToUpdate.Spec.Buckets = &newBuckets
		updatedSA, err := client.Storage().BucketServiceAccounts().Update(ctx, "example-sa-single", saToUpdate)
		if err != nil {
			log.Printf("   Warning: Update failed (may not be supported): %v", err)
		} else {
			bucketCount := 0
			if updatedSA.Spec.Buckets != nil {
				bucketCount = len(*updatedSA.Spec.Buckets)
			}
			fmt.Printf("   ✓ Updated: %s (now has %d buckets)\n", *updatedSA.Metadata.Name, bucketCount)
		}
	}

	// Example 7: Delete a service account
	fmt.Println("\n7. Deleting a service account...")
	err = client.Storage().BucketServiceAccounts().Delete(ctx, "example-sa-individual")
	if err != nil {
		return fmt.Errorf("failed to delete service account: %w", err)
	}
	fmt.Println("   ✓ Deleted example-sa-individual")

	return nil
}
