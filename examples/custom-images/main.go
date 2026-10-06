// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Register an uploaded image and create a disk from it.
// Requires the custom-image API, normal evroc SDK authentication/configuration,
// and a bucket object plus a bucket service account authorized to read it.
//
// Required environment: CUSTOM_IMAGE_NAME, CUSTOM_IMAGE_BUCKET_REF,
// CUSTOM_IMAGE_BUCKET_ACCOUNT_REF, CUSTOM_IMAGE_OBJECT_PATH, CUSTOM_IMAGE_OBJECT_VERSION,
// CUSTOM_IMAGE_DISK_NAME, CUSTOM_IMAGE_DISK_SIZE_GB, CUSTOM_IMAGE_ZONE.
// Bucket references are full resource IDs. Names must be unused. This example
// creates billable resources and leaves them in place for inspection.
//
// Run from the repository root: go run ./examples/custom-images
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/compute"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	names := []string{"CUSTOM_IMAGE_NAME", "CUSTOM_IMAGE_BUCKET_REF", "CUSTOM_IMAGE_BUCKET_ACCOUNT_REF", "CUSTOM_IMAGE_OBJECT_PATH", "CUSTOM_IMAGE_OBJECT_VERSION", "CUSTOM_IMAGE_DISK_NAME", "CUSTOM_IMAGE_DISK_SIZE_GB", "CUSTOM_IMAGE_ZONE"}
	values := make(map[string]string, len(names))
	for _, name := range names {
		value := os.Getenv(name)
		if value == "" {
			return fmt.Errorf("%s is required", name)
		}
		values[name] = value
	}
	size, err := strconv.ParseInt(values["CUSTOM_IMAGE_DISK_SIZE_GB"], 10, 32)
	if err != nil || size <= 0 {
		return fmt.Errorf("CUSTOM_IMAGE_DISK_SIZE_GB must be a positive int32")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	image, err := compute.NewCustomDiskImageBuilder(values["CUSTOM_IMAGE_NAME"], values["CUSTOM_IMAGE_BUCKET_REF"], values["CUSTOM_IMAGE_BUCKET_ACCOUNT_REF"], values["CUSTOM_IMAGE_OBJECT_PATH"]).
		WithDefaultDiskSizeGB(int32(size)).WithObjectVersion(values["CUSTOM_IMAGE_OBJECT_VERSION"]).
		Create(ctx, client.Compute().CustomDiskImages())
	if err != nil {
		return fmt.Errorf("register image: %w", err)
	}
	fmt.Printf("Registered image %s\n", image.Metadata.Id)
	if _, err := client.Compute().CustomDiskImages().WaitForReady(ctx, image.Metadata.Id, 2*time.Minute); err != nil {
		return fmt.Errorf("wait for image registration: %w", err)
	}
	disk, err := compute.NewDiskBuilder(values["CUSTOM_IMAGE_DISK_NAME"]).
		WithCustomImage(client.Compute().CustomDiskImageRef(image.Metadata.Id)).
		WithSizeGB(int32(size)).WithZone(values["CUSTOM_IMAGE_ZONE"]).Create(ctx, client.Compute().Disks())
	if err != nil {
		return fmt.Errorf("create disk: %w", err)
	}
	if _, err := client.Compute().Disks().WaitForReady(ctx, disk.Metadata.Id, 5*time.Minute); err != nil {
		return fmt.Errorf("wait for disk: %w", err)
	}
	fmt.Printf("Disk %s is ready; the image and disk remain for inspection.\n", disk.Metadata.Id)
	return nil
}
