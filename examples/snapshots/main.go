// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Disk Snapshots Example
//
// Demonstrates the snapshot lifecycle:
//  1. Create a disk with an OS image
//  2. Take a snapshot of the disk
//  3. Create a new, larger disk from the snapshot (restore + resize)
//  4. Clean up all resources
//
// Usage:
//
//	go run main.go
//	go run main.go destroy
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/compute"
)

const (
	srcDiskName      = "sdk-snap-src"
	snapshotName     = "sdk-snap-backup"
	restoredDiskName = "sdk-snap-restored"
	zone             = "a"
)

func main() {
	ctx := context.Background()
	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 && os.Args[1] == "destroy" {
		destroy(ctx, client)
		return
	}

	fmt.Println("=== Disk Snapshot Example ===")

	// Step 1: Create a source disk
	fmt.Println("Step 1: Creating source disk...")
	srcDisk, err := compute.NewDiskBuilder(srcDiskName).
		WithImage(string(compute.DiskImageUbuntuMinimal2404)).
		WithSizeGB(20).
		WithZone(zone).
		Create(ctx, client.Compute().Disks())
	if err != nil {
		log.Fatalf("Failed to create disk: %v", err)
	}

	fmt.Println("  Waiting for disk...")
	srcDisk, err = client.Compute().Disks().WaitForReady(ctx, srcDiskName, 3*time.Minute)
	if err != nil {
		log.Fatalf("Disk never became ready: %v", err)
	}
	fmt.Printf("  Ready: %s (%dGB)\n\n",
		srcDisk.Metadata.Id, srcDisk.Spec.DiskSize.Amount)

	// Step 2: Take a snapshot
	fmt.Println("Step 2: Taking snapshot...")
	_, err = compute.NewSnapshotBuilder(snapshotName).
		WithDiskRef(string(srcDisk.Ref())).
		Create(ctx, client.Compute().Snapshots())
	if err != nil {
		log.Fatalf("Failed to create snapshot: %v", err)
	}

	fmt.Println("  Waiting for snapshot...")
	snap, err := client.Compute().Snapshots().WaitForReady(ctx, snapshotName, 5*time.Minute)
	if err != nil {
		log.Fatalf("Snapshot never became ready: %v", err)
	}
	fmt.Printf("  Ready: %s\n\n", snap.Metadata.Id)

	// Step 3: Restore to a new, larger disk
	fmt.Println("Step 3: Restoring snapshot to a larger disk (20GB -> 50GB)...")
	_, err = compute.NewDiskBuilder(restoredDiskName).
		WithSnapshot(client.Compute().SnapshotRef(snapshotName)).
		WithSizeGB(50).
		WithZone(zone).
		Create(ctx, client.Compute().Disks())
	if err != nil {
		log.Fatalf("Failed to create restored disk: %v", err)
	}

	fmt.Println("  Waiting for restored disk...")
	restored, err := client.Compute().Disks().WaitForReady(ctx, restoredDiskName, 3*time.Minute)
	if err != nil {
		log.Fatalf("Restored disk never became ready: %v", err)
	}
	fmt.Printf("  Ready: %s (%dGB, restored from snapshot)\n\n",
		restored.Metadata.Id, restored.Spec.DiskSize.Amount)

	// Step 4: List all snapshots
	fmt.Println("Step 4: Listing snapshots...")
	snapshots, err := client.Compute().Snapshots().List(ctx)
	if err != nil {
		log.Fatalf("Failed to list snapshots: %v", err)
	}
	for _, s := range snapshots.Items {
		ready := compute.IsSnapshotReady(&s)
		fmt.Printf("  %s (ready: %v)\n", s.Metadata.Id, ready)
	}

	fmt.Println("\n=== Done ===")
	fmt.Println("Run 'go run main.go destroy' to clean up")
}

func destroy(ctx context.Context, client *evroc.Client) {
	fmt.Println("=== Cleaning up ===")
	c := client.Compute()

	c.Disks().Delete(ctx, restoredDiskName)
	fmt.Printf("  Deleted: %s\n", restoredDiskName)

	c.Snapshots().Delete(ctx, snapshotName)
	c.Snapshots().WaitForDeleted(ctx, snapshotName, 2*time.Minute)
	fmt.Printf("  Deleted: %s\n", snapshotName)

	c.Disks().Delete(ctx, srcDiskName)
	fmt.Printf("  Deleted: %s\n", srcDiskName)

	fmt.Println("Done")
}
