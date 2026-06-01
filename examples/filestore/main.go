// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// FileStore example
//
// Demonstrates creating an NFS file store, waiting for it to become available,
// printing the NFS connection details, and then cleaning up.
//
// This example is idempotent — it can be re-run safely. If the file store
// already exists, it picks up where a previous run left off.
//
// # Running
//
//	export EVROC_USERNAME=<username>
//	export EVROC_PASSWORD=<password>
//	export EVROC_PROJECT=<project-id>
//	export EVROC_REGION=se-sto
//	export EVROC_ORGANIZATION=<org-id>
//
//	go run main.go
//
// # What This Example Shows
//
//  1. Creating a file store with the builder pattern (or reusing an existing one)
//  2. Waiting for the file store to become available
//  3. Reading NFS connection details (endpoint, export path)
//  4. Listing all file stores
//  5. Deleting the file store and waiting for cleanup
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/storage"
)

const (
	fileStoreName = "sdk-example-nfs"
	zone          = "a"

	availableTimeout = 3 * time.Minute
	deletedTimeout   = 2 * time.Minute
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== FileStore Example ===")
	fmt.Println()

	// 1. Create a file store (or pick up an existing one)
	fmt.Println("1. Creating file store...")
	_, err = storage.NewFileStoreBuilder(fileStoreName, zone).
		WithLabels(map[string]string{"purpose": "sdk-example"}).
		Create(ctx, client.Storage().FileStores())
	if err != nil {
		if errors.Is(err, evroc.ErrConflict) {
			fmt.Printf("   Already exists, reusing\n")
		} else {
			log.Fatalf("Failed to create file store: %v", err)
		}
	} else {
		fmt.Printf("   Created: %s\n", fileStoreName)
	}

	// 2. Wait for it to become available
	fmt.Println("\n2. Waiting for file store to become available...")
	fs, err := client.Storage().FileStores().WaitForAvailable(ctx, fileStoreName, availableTimeout)
	if err != nil {
		log.Fatalf("File store did not become available: %v", err)
	}
	fmt.Printf("   File store is available\n")

	// 3. Print NFS connection details
	fmt.Println("\n3. NFS connection details:")
	fmt.Printf("   Endpoint:    %s\n", fs.Status.Nfs.Endpoint)
	fmt.Printf("   Export Path: %s\n", fs.Status.Nfs.ExportPath)
	fmt.Printf("   NFS Version: %s\n", fs.Status.Nfs.Version)
	fmt.Printf("   Zone:        %s\n", fs.Status.Placement.Zone)
	fmt.Println()
	fmt.Printf("   Mount command:\n")
	fmt.Printf("   sudo mount -t nfs -o vers=4.1 %s:%s /mnt/filestore\n",
		fs.Status.Nfs.Endpoint, fs.Status.Nfs.ExportPath)

	// 4. List all file stores
	fmt.Println("\n4. Listing all file stores...")
	list, err := client.Storage().FileStores().List(ctx)
	if err != nil {
		log.Fatalf("Failed to list file stores: %v", err)
	}
	fmt.Printf("   Found %d file store(s):\n", len(list.Items))
	for _, item := range list.Items {
		status := "unknown"
		if item.Status.Status != nil {
			status = string(*item.Status.Status)
		}
		fmt.Printf("   - %s (zone: %s, status: %s)\n",
			item.Metadata.Id, item.Spec.Placement.Zone, status)
	}

	// 5. Delete and wait for cleanup
	fmt.Println("\n5. Deleting file store...")
	if err := client.Storage().FileStores().Delete(ctx, fileStoreName); err != nil {
		if errors.Is(err, evroc.ErrNotFound) {
			fmt.Printf("   Already deleted\n")
		} else {
			log.Fatalf("Failed to delete file store: %v", err)
		}
	} else {
		fmt.Printf("   Delete requested, waiting for cleanup...\n")
		if err := client.Storage().FileStores().WaitForDeleted(ctx, fileStoreName, deletedTimeout); err != nil {
			log.Fatalf("File store was not deleted in time: %v", err)
		}
		fmt.Printf("   File store deleted\n")
	}

	fmt.Println("\n=== FileStore Example Complete ===")
}
