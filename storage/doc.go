// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package storage provides access to the evroc Storage API.
//
// The Storage API enables you to manage S3-compatible object storage buckets,
// service accounts, and access credentials in the evroc Cloud Platform.
//
// # Resources
//
// The storage package provides access to the following resources:
//
//   - Buckets: S3-compatible object storage containers
//   - Bucket Service Accounts: Credentials for programmatic bucket access
//   - Bucket Service Account Secrets: Access keys and secret keys
//
// # Getting Started
//
// Create a client and list buckets:
//
//	client, err := evroc.NewFromEnv(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	buckets, err := client.Storage().Buckets().List(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Creating Buckets
//
// Create buckets with optional object locking and versioning:
//
//	bucket, err := client.Storage().Buckets().Create(ctx,
//	    storage.NewBucketBuilder("my-bucket").
//	        WithObjectRetentionMode(storage.Versioned).
//	        Build(),
//	)
//
// Available retention modes: storage.Disabled, storage.Suspended, storage.Versioned, storage.Locking
//
// For buckets with locking enabled:
//
//	bucket, err := client.Storage().Buckets().Create(ctx,
//	    storage.NewBucketBuilder("my-bucket").
//	        WithObjectRetentionMode(storage.Locking).
//	        WithDefaultObjectLocking(storage.Immutable, 30).
//	        Build(),
//	)
//
// Available locking modes: storage.Immutable (COMPLIANCE), storage.Soft (GOVERNANCE)
//
// # Updating Buckets
//
// Modify bucket settings using update builders:
//
//	bucket, err := storage.UpdateBucket("my-bucket", client.Storage().Buckets()).
//	    SetRetentionMode(storage.RetentionModeGovernance).
//	    Apply(ctx)
//
// # Service Accounts
//
// Create service accounts for programmatic access:
//
//	sa, err := client.Storage().BucketServiceAccounts().Create(ctx,
//	    storage.NewBucketServiceAccountBuilder("app-sa").
//	        WithDescription("Application service account").
//	        WithBucket("my-bucket").
//	        Build(),
//	)
//
// Grant a service account access to multiple buckets:
//
//	sa, err := storage.UpdateBucketServiceAccount("app-sa", client.Storage().BucketServiceAccounts()).
//	    AddBucket("bucket-1").
//	    AddBucket("bucket-2").
//	    Apply(ctx)
//
// # Access Credentials
//
// Create access keys for service accounts:
//
//	secret, err := client.Storage().BucketServiceAccountSecrets().Create(ctx,
//	    storage.NewBucketServiceAccountSecretBuilder("app-key").
//	        WithDescription("Production credentials").
//	        ForServiceAccount("app-sa").
//	        Build(),
//	)
//
//	// Access credentials
//	accessKey := secret.Status.AccessKey
//	secretKey := secret.Status.SecretKey
//
// # S3 Client Configuration
//
// Use the credentials with an S3-compatible client. MinIO SDK is recommended
// for its simpler API and better handling of S3-compatible services:
//
//	import "github.com/minio/minio-go/v7"
//	import "github.com/minio/minio-go/v7/pkg/credentials"
//
//	s3Client, err := minio.New(client.Storage().GetS3Endpoint(), &minio.Options{
//	    Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
//	    Secure: true,
//	    Region: client.DefaultRegion(),
//	})
//
// AWS SDK Go is also supported:
//
//	import "github.com/aws/aws-sdk-go/aws"
//	import "github.com/aws/aws-sdk-go/aws/credentials"
//	import "github.com/aws/aws-sdk-go/aws/session"
//	import "github.com/aws/aws-sdk-go/service/s3"
//
//	endpoint := "https://" + client.Storage().GetS3Endpoint()
//	sess := session.Must(session.NewSession(&aws.Config{
//	    Endpoint:         aws.String(endpoint),
//	    Region:           aws.String(client.DefaultRegion()),
//	    Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
//	    S3ForcePathStyle: aws.Bool(true),
//	}))
//
//	s3Client := s3.New(sess)
//
// # Context Support
//
// All operations support context for cancellation and timeouts:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	bucket, err := client.Storage().Buckets().Get(ctx, "my-bucket")
package storage
