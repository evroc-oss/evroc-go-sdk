// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

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
//	        WithDescription("Application data").
//	        WithVersioning(true).
//	        WithObjectLocking(storage.LockingModeCompliance).
//	        Build(),
//	)
//
// Note: Use type aliases for cleaner code:
//   - storage.RetentionMode instead of storage.BucketSpecObjectRetentionMode
//   - storage.LockingMode instead of storage.BucketSpecDefaultObjectLockingMode
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
// Use the credentials with an S3 client:
//
//	import "github.com/aws/aws-sdk-go/aws"
//	import "github.com/aws/aws-sdk-go/aws/credentials"
//	import "github.com/aws/aws-sdk-go/aws/session"
//	import "github.com/aws/aws-sdk-go/service/s3"
//
//	sess := session.Must(session.NewSession(&aws.Config{
//	    Endpoint:         aws.String("https://s3.se-sto.cloud.evroc.com"),
//	    Region:           aws.String("se-sto"),
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
