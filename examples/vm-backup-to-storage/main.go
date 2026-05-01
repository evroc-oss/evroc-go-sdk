// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// VM Backup to Object Storage Example
//
// This example demonstrates using the Evroc SDK to:
//  1. Create an S3-compatible bucket
//  2. Create a service account with S3 credentials
//  3. Use the credentials to upload/download files via S3 API (using MinIO Go SDK)
//  4. Demonstrate object versioning
//  5. Clean up all resources
//
// # Running
//
// This example uses a separate Go module to keep MinIO SDK out of the main SDK dependencies.
//
//	# From repo root:
//	go run examples/vm-backup-to-storage/main.go
//
//	# Or from the example directory:
//	cd examples/vm-backup-to-storage
//	go run main.go
//
// # What it does
//
//   - Creates a bucket and service account via Evroc SDK
//   - Retrieves S3 credentials programmatically (shows full credentials in output)
//   - Uploads a 20MB file and measures upload speed (MB/s)
//   - Downloads the file and measures download speed (MB/s)
//   - Verifies data integrity
//   - Uploads a second version of the same file (different content)
//   - Lists all object versions to demonstrate versioning
//   - Deletes all objects from bucket
//   - Cleans up service account and bucket
//
// # Dependencies
//
// This example has its own go.mod to isolate the MinIO SDK dependency from the main SDK.
// The root go.work file manages both modules together.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/storage"
)

const (
	bucketName         = "sdk-storage"
	serviceAccountName = "sdk-sa"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Cleanup at start
	cleanupResources(ctx, client, nil)

	// 1. Create bucket
	fmt.Println("1. Creating bucket...")
	_, err = client.Storage().Buckets().Create(ctx,
		storage.NewBucketBuilder(bucketName).Build())
	if err != nil {
		if strings.Contains(err.Error(), "AlreadyExists") {
			log.Fatalf("Bucket %s already exists. Please delete it first or wait for cleanup to complete.", bucketName)
		} else {
			log.Fatalf("Failed to create bucket: %v", err)
		}
	}
	fmt.Printf("   ✓ Created bucket: %s\n", bucketName)

	// 2. Create service account
	fmt.Println("\n2. Creating service account...")
	_, err = client.Storage().BucketServiceAccounts().Create(ctx,
		storage.NewBucketServiceAccountBuilder(serviceAccountName).
			WithBucket(bucketName).
			Build())
	if err != nil {
		if strings.Contains(err.Error(), "AlreadyExists") {
			log.Fatalf("Service account %s already exists. Please delete it first or wait for cleanup to complete.", serviceAccountName)
		} else {
			log.Fatalf("Failed to create service account: %v", err)
		}
	}
	fmt.Printf("   ✓ Created service account: %s\n", serviceAccountName)

	// 3. Wait for credentials and get them
	fmt.Println("\n3. Waiting for S3-compatible credentials...")
	creds, err := client.Storage().BucketServiceAccounts().WaitForCredentials(ctx, serviceAccountName, 60*time.Second)
	if err != nil {
		log.Fatalf("Failed waiting for credentials: %v", err)
	}

	accessKey := creds.AccessKeyID
	secretKey := creds.SecretAccessKey
	fmt.Printf("   ✓ Access Key ID: %s\n", accessKey)
	fmt.Printf("   ✓ Secret Access Key: %s\n", secretKey)

	// 5. Create S3 client using MinIO SDK
	fmt.Println("\n5. Creating S3 client...")
	s3Client, err := minio.New(client.Storage().GetS3Endpoint(), &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
		Region: client.DefaultRegion(),
	})
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}
	fmt.Println("   ✓ S3 client ready")

	// Defer cleanup with S3 client for object deletion
	defer cleanupResources(ctx, client, s3Client)

	// 6. Upload a 20MB file
	fileSize := int64(20 * 1024 * 1024) // 20 MB
	testData := make([]byte, fileSize)
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	fmt.Printf("\n6. Uploading test file (%.2f MB)...\n", float64(fileSize)/(1024*1024))
	start := time.Now()
	_, err = s3Client.PutObject(ctx, bucketName, "test-file.dat",
		bytes.NewReader(testData), fileSize,
		minio.PutObjectOptions{ContentType: "application/octet-stream"})
	uploadTime := time.Since(start)
	if err != nil {
		log.Fatalf("Failed to upload: %v", err)
	}
	fmt.Printf("   ✓ Uploaded in %v (%.2f MB/s)\n",
		uploadTime, float64(fileSize)/(1024*1024)/uploadTime.Seconds())

	// 7. Download the file
	fmt.Println("\n7. Downloading test file...")
	start = time.Now()
	obj, err := s3Client.GetObject(ctx, bucketName, "test-file.dat", minio.GetObjectOptions{})
	if err != nil {
		log.Fatalf("Failed to download: %v", err)
	}
	downloadedData, err := io.ReadAll(obj)
	obj.Close()
	downloadTime := time.Since(start)
	if err != nil {
		log.Fatalf("Failed to read downloaded data: %v", err)
	}
	downloadedSize := float64(len(downloadedData)) / (1024 * 1024)
	fmt.Printf("   ✓ Downloaded %.2f MB in %v (%.2f MB/s)\n",
		downloadedSize, downloadTime, downloadedSize/downloadTime.Seconds())

	// 8. Verify content
	fmt.Println("\n8. Verifying content...")
	if bytes.Equal(testData, downloadedData) {
		fmt.Println("   ✓ Content matches!")
	} else {
		log.Fatal("   ✗ Content mismatch!")
	}

	// 9. Upload a new version of the file (modified content)
	fileSize2 := int64(20 * 1024 * 1024) // 20 MB
	testData2 := make([]byte, fileSize2)
	for i := range testData2 {
		testData2[i] = byte((i + 128) % 256) // Different pattern
	}
	fmt.Printf("\n9. Uploading new version (%.2f MB)...\n", float64(fileSize2)/(1024*1024))
	_, err = s3Client.PutObject(ctx, bucketName, "test-file.dat",
		bytes.NewReader(testData2), fileSize2,
		minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		log.Fatalf("Failed to upload version 2: %v", err)
	}
	fmt.Println("   ✓ Uploaded version 2")

	// 10. List object versions
	fmt.Println("\n10. Listing object versions...")
	objectCh := s3Client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:       "test-file.dat",
		WithVersions: true,
	})
	versionCount := 0
	for object := range objectCh {
		if object.Err != nil {
			log.Fatalf("Error listing objects: %v", object.Err)
		}
		versionCount++
		sizeMB := float64(object.Size) / (1024 * 1024)
		fmt.Printf("   - Version %d: %s (%.2f MB, %v)\n",
			versionCount, object.VersionID, sizeMB, object.LastModified.Format(time.RFC3339))
	}
	fmt.Printf("   ✓ Found %d version(s)\n", versionCount)

	fmt.Println("\n✓ All operations completed successfully")
}

func cleanupResources(ctx context.Context, client *evroc.Client, s3Client *minio.Client) {
	// Delete all objects from bucket first (if s3Client is available)
	if s3Client != nil {
		_, err := client.Storage().Buckets().Get(ctx, bucketName)
		if err == nil {
			fmt.Printf("Deleting objects from bucket %s...\n", bucketName)
			objectCh := s3Client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
				Recursive:    true,
				WithVersions: true,
			})
			for object := range objectCh {
				if object.Err != nil {
					fmt.Printf("Warning: error listing objects: %v\n", object.Err)
					break
				}
				if err := s3Client.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{
					VersionID: object.VersionID,
				}); err != nil {
					fmt.Printf("Warning: failed to delete object %s: %v\n", object.Key, err)
				}
			}
			fmt.Println("Objects deleted")
		}
	}

	// Delete service account
	_, err := client.Storage().BucketServiceAccounts().Get(ctx, serviceAccountName)
	if err == nil {
		fmt.Printf("Deleting service account %s...\n", serviceAccountName)
		if err := client.Storage().BucketServiceAccounts().Delete(ctx, serviceAccountName); err != nil {
			fmt.Printf("Warning: failed to delete service account: %v\n", err)
			return
		}
		if err := client.Storage().BucketServiceAccounts().WaitForDeleted(ctx, serviceAccountName, 60*time.Second); err != nil {
			fmt.Printf("Warning: timeout waiting for service account deletion: %v\n", err)
			return
		}
		fmt.Println("Service account deleted")
	}

	// Delete bucket
	_, err = client.Storage().Buckets().Get(ctx, bucketName)
	if err == nil {
		fmt.Printf("Deleting bucket %s...\n", bucketName)
		if err := client.Storage().Buckets().Delete(ctx, bucketName); err != nil {
			fmt.Printf("Warning: failed to delete bucket: %v\n", err)
			return
		}
		if err := client.Storage().Buckets().WaitForDeleted(ctx, bucketName, 60*time.Second); err != nil {
			fmt.Printf("Warning: timeout waiting for bucket deletion: %v\n", err)
			return
		}
		fmt.Println("Bucket deleted")
	}
}
