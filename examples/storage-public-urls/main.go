// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Storage Presigned URLs Example
//
// Demonstrates S3-compatible object storage with presigned URLs using the evroc SDK.
//
// # Running
//
//	export EVROC_TOKEN=<token>
//	export EVROC_PROJECT=<project-id>
//	export EVROC_REGION=se-sto
//
//	go run main.go
//
// # What This Example Shows
//
//  1. Creating bucket and service account
//  2. Getting S3 credentials with WaitForCredentials
//  3. Uploading files with proper MIME types
//  4. Creating presigned URLs for browser downloads (no credentials needed)
//  5. Multipart upload for large files
//  6. JavaScript upload pattern
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
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
	bucketName         = "sdk-urls"
	serviceAccountName = "sdk-urls-sa"
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

	// 3. Wait for credentials
	fmt.Println("\n3. Waiting for S3-compatible credentials...")
	creds, err := client.Storage().BucketServiceAccounts().WaitForCredentials(ctx, serviceAccountName, 60*time.Second)
	if err != nil {
		log.Fatalf("Failed waiting for credentials: %v", err)
	}

	accessKey := creds.AccessKeyID
	secretKey := creds.SecretAccessKey
	fmt.Println("   ✓ Credentials ready")

	// 4. Create S3 client
	fmt.Println("\n4. Creating S3 client...")
	s3Client, err := minio.New(client.Storage().GetS3Endpoint(), &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
		Region: client.DefaultRegion(),
	})
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}
	fmt.Println("   ✓ S3 client ready")

	defer cleanupResources(ctx, client, s3Client)

	// 5. Upload files with different MIME types
	fmt.Println("\n5. Uploading files with proper MIME types...")
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x00, // placeholder bytes for demo content
	}

	files := []struct {
		key         string
		content     []byte
		contentType string
	}{
		{"report.txt", []byte("Sample report"), "text/plain"},
		{"evroc-logo.png", pngData, "image/png"},
		{"data.json", []byte(`{"status":"ok"}`), "application/json"},
	}

	for _, f := range files {
		_, err = s3Client.PutObject(ctx, bucketName, f.key,
			bytes.NewReader(f.content), int64(len(f.content)),
			minio.PutObjectOptions{ContentType: f.contentType})
		if err != nil {
			log.Fatalf("Failed to upload %s: %v", f.key, err)
		}
		fmt.Printf("   ✓ Uploaded: %s (%.2f KB)\n", f.key, float64(len(f.content))/1024)
	}

	// 6. Presigned URL for download
	fmt.Println("\n6. Presigned URL (works in browser without credentials)...")
	presignedURL, err := s3Client.PresignedGetObject(ctx, bucketName, "report.txt", 15*time.Minute, nil)
	if err != nil {
		log.Fatalf("Failed to generate presigned URL: %v", err)
	}
	fmt.Printf("   %s\n", presignedURL.String())

	// Test download via presigned URL
	resp, err := http.Get(presignedURL.String())
	if err != nil {
		log.Fatalf("Failed to download: %v", err)
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(resp.Body)
	fmt.Printf("   ✓ Downloaded %d bytes: %s\n", len(content), string(content))

	// 7. Multipart upload (10MB file in 5MB chunks)
	// Note: S3 response-* query parameters (like response-content-disposition) are not supported
	fmt.Println("\n7. Multipart upload (10MB file)...")
	largeFileSize := int64(10 * 1024 * 1024) // 10 MB
	largeData := make([]byte, largeFileSize)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	start := time.Now()
	_, err = s3Client.PutObject(ctx, bucketName, "large-file.bin",
		bytes.NewReader(largeData), largeFileSize,
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
			PartSize:    5 * 1024 * 1024, // 5MB parts (S3 minimum)
		})
	if err != nil {
		log.Fatalf("Failed to upload large file: %v", err)
	}
	duration := time.Since(start)
	fmt.Printf("   ✓ Uploaded %.2f MB in %v (%.2f MB/s)\n",
		float64(largeFileSize)/(1024*1024),
		duration,
		float64(largeFileSize)/(1024*1024)/duration.Seconds())
	fmt.Printf("   ✓ MinIO SDK automatically split into 2 parts of 5MB each\n")

	fmt.Println("\n✓ Example complete!")
	fmt.Println("\nPresigned URLs above are valid for 15 minutes.")

	// Show how to use presigned URLs from JavaScript
	fmt.Println("\n--- JavaScript Upload Example ---")
	fmt.Println("To upload from a browser, generate a presigned PUT URL:")
	uploadURL, _ := s3Client.PresignedPutObject(ctx, bucketName, "user-upload.txt", 15*time.Minute)
	fmt.Printf("Presigned PUT URL: %s\n\n", uploadURL.String())
	fmt.Println("Then in JavaScript:")
	fmt.Println("  fetch(presignedPutURL, {")
	fmt.Println("    method: 'PUT',")
	fmt.Println("    body: file,")
	fmt.Println("    headers: { 'Content-Type': 'text/plain' }")
	fmt.Println("  })")
	fmt.Println("---")

	fmt.Println("\nWaiting 30 seconds before cleanup (press Ctrl+C to keep resources)...")
	time.Sleep(30 * time.Second)
}

func cleanupResources(ctx context.Context, client *evroc.Client, s3Client *minio.Client) {
	// Delete objects if S3 client available
	if s3Client != nil {
		objectCh := s3Client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
			Recursive:    true,
			WithVersions: true,
		})
		for object := range objectCh {
			if object.Err != nil {
				break
			}
			s3Client.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{
				VersionID: object.VersionID,
			})
		}
	}

	// Delete service account and wait for deletion
	client.Storage().BucketServiceAccounts().Delete(ctx, serviceAccountName)
	client.Storage().BucketServiceAccounts().WaitForDeleted(ctx, serviceAccountName, 30*time.Second)

	// Delete bucket
	client.Storage().Buckets().Delete(ctx, bucketName)
}
