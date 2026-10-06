// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package compute_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/compute"
	"github.com/evroc-oss/evroc-go-sdk/internal/e2etest"
	"github.com/evroc-oss/evroc-go-sdk/storage"
)

// TestE2E_CustomDiskImage_Lifecycle uploads a small raw image to a new bucket,
// registers it, and creates a disk from it. Cleanups run in reverse order, so the
// disk is gone before the image, and the image before its bucket.
// Outside production, set E2E_S3_ENDPOINT to the region's S3 host.
func TestE2E_CustomDiskImage_Lifecycle(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client := e2etest.NewClient(t)
	project, region := client.DefaultProject(), client.DefaultRegion()
	name := e2etest.RandomName("cdi")

	// Versioned, so the registration can pin an object version.
	if _, err := storage.NewBucketBuilder(name).WithObjectRetentionMode(storage.Versioned).Create(ctx, client.Storage().Buckets()); err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}
	cleanup(t, "bucket", func() error { return client.Storage().Buckets().Delete(ctx, name) })
	if _, err := client.Storage().Buckets().WaitForReady(ctx, name, 2*time.Minute); err != nil {
		t.Fatalf("bucket never became ready: %v", err)
	}

	if _, err := storage.NewBucketServiceAccountBuilder(name).WithBucket(name).Create(ctx, client.Storage().BucketServiceAccounts()); err != nil {
		t.Fatalf("failed to create bucket service account: %v", err)
	}
	cleanup(t, "bucket service account", func() error { return client.Storage().BucketServiceAccounts().Delete(ctx, name) })
	creds, err := client.Storage().BucketServiceAccounts().WaitForCredentials(ctx, name, 2*time.Minute)
	if err != nil {
		t.Fatalf("credentials never became ready: %v", err)
	}

	// A raw image of zeros is a valid disk image; it only has to import, not boot.
	object := e2etest.S3Object{
		Endpoint: e2etest.S3Endpoint(client.Storage().GetS3Endpoint()), Region: region,
		AccessKey: creds.AccessKeyID, SecretKey: creds.SecretAccessKey,
		Bucket: name, Key: "images/" + name + ".raw",
	}
	var version string
	for attempt := 1; attempt <= 6; attempt++ { // fresh credentials take a moment to reach S3
		if version, err = object.Put(ctx, make([]byte, 1<<20)); err == nil {
			break
		}
		time.Sleep(10 * time.Second)
	}
	if err != nil || version == "" {
		t.Fatalf("failed to upload versioned image object (version %q): %v", version, err)
	}
	cleanup(t, "image object", func() error { return object.Delete(ctx, version) })

	bucketRef := fmt.Sprintf("/storage/projects/%s/regions/%s/buckets/%s", project, region, name)
	saRef := fmt.Sprintf("/storage/projects/%s/regions/%s/bucketServiceAccounts/%s", project, region, name)
	if _, err := compute.NewCustomDiskImageBuilder(name, bucketRef, saRef, object.Key).
		WithDefaultDiskSizeGB(10).WithObjectVersion(version).
		Create(ctx, client.Compute().CustomDiskImages()); err != nil {
		t.Fatalf("failed to register custom disk image: %v", err)
	}
	cleanup(t, "custom disk image", func() error {
		if err := client.Compute().CustomDiskImages().Delete(ctx, name); err != nil {
			return err
		}
		return client.Compute().CustomDiskImages().WaitForDeleted(ctx, name, 3*time.Minute)
	})
	image, err := client.Compute().CustomDiskImages().WaitForReady(ctx, name, 3*time.Minute)
	if err != nil {
		t.Fatalf("custom disk image never became ready: %v", err)
	}
	// The platform must accept and return the source as the SDK sends it.
	if s := image.Spec.Source; s.BucketRef != bucketRef || s.ServiceAccountRef != saRef || s.Version == nil || *s.Version != version {
		t.Errorf("unexpected source: %+v", s)
	}

	// No size, so the image default applies.
	if _, err := compute.NewDiskBuilder(name).WithCustomImage(image.Ref()).WithZone(e2etest.TestDiskZone).Create(ctx, client.Compute().Disks()); err != nil {
		t.Fatalf("failed to create disk from custom image: %v", err)
	}
	cleanup(t, "disk", func() error {
		if err := client.Compute().Disks().Delete(ctx, name); err != nil {
			return err
		}
		return client.Compute().Disks().WaitForDeleted(ctx, name, 5*time.Minute)
	})
	if _, err := client.Compute().Disks().WaitForReady(ctx, name, 10*time.Minute); err != nil {
		t.Fatalf("disk from custom image never became ready: %v", err)
	}
}

func cleanup(t *testing.T, what string, fn func() error) {
	t.Cleanup(func() {
		if err := fn(); err != nil {
			t.Errorf("failed to clean up %s: %v", what, err)
		}
	})
}
