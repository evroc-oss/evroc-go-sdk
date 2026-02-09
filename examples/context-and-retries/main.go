// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package main demonstrates best practices for context usage, retry handling, and waiter configuration.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/compute"
)

func main() {
	fmt.Println("=== Context and Retry Best Practices ===")
	fmt.Println()

	// Example 1: Using context with timeout
	demonstrateContextTimeout()

	// Example 2: Using context with cancellation
	demonstrateContextCancellation()

	// Example 3: Customizing waiter behavior
	demonstrateCustomWaiters()

	// Example 4: Handling retries
	demonstrateRetryBehavior()

	fmt.Println("\n=== All Examples Complete ===")
}

// demonstrateContextTimeout shows proper use of context.WithTimeout.
func demonstrateContextTimeout() {
	fmt.Println("--- Example 1: Context with Timeout ---")
	fmt.Println("Best practice: Always set timeouts for operations that might hang")
	fmt.Println()

	// Create a context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Printf("Failed to create client: %v (this is expected if credentials not set)", err)
		fmt.Println()
		return
	}

	// List VMs with timeout
	fmt.Println("Listing VMs with 30s timeout...")
	start := time.Now()
	vms, err := client.Compute().VirtualMachines().List(ctx)
	elapsed := time.Since(start)

	if err != nil {
		log.Printf("Operation failed after %v: %v", elapsed, err)
	} else {
		fmt.Printf("✓ Listed %d VMs in %v\n", len(vms.Items), elapsed)
	}
	fmt.Println()
}

// demonstrateContextCancellation shows proper handling of user cancellation.
func demonstrateContextCancellation() {
	fmt.Println("--- Example 2: Context with Cancellation ---")
	fmt.Println("Best practice: Respect SIGINT/SIGTERM for graceful shutdown")
	fmt.Println()

	// Create a context that cancels on Ctrl-C or SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Println("Context will cancel on Ctrl-C (SIGINT) or SIGTERM")
	fmt.Println("All SDK operations will stop when context is cancelled")
	fmt.Println()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Printf("Failed to create client: %v (this is expected if credentials not set)", err)
		fmt.Println()
		return
	}

	// Example: Create a disk that respects cancellation
	fmt.Println("Example: Creating disk with cancellation support...")
	disk := compute.NewDiskBuilder("cancellable-disk").
		WithImage(string(compute.DiskImageUbuntuMinimal2404)).
		WithSizeGB(20).
		WithZone("a").
		Build()

	// This will respect context cancellation during the API call
	_, err = client.Compute().Disks().Create(ctx, disk)
	if err != nil {
		fmt.Printf("Note: Operation cancelled or failed: %v\n", err)
	}
	fmt.Println()
}

// demonstrateCustomWaiters shows how to customize waiter behavior.
func demonstrateCustomWaiters() {
	fmt.Println("--- Example 3: Customizing Waiter Behavior ---")
	fmt.Println("Waiters now support context cancellation and custom polling intervals")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Printf("Failed to create client: %v (this is expected if credentials not set)", err)
		fmt.Println()
		return
	}

	// Example 3a: Default waiter behavior
	fmt.Println("3a. Default waiter (exponential backoff, 2s initial, 30s max):")
	disk := compute.NewDiskBuilder("demo-disk-default").
		WithImage(string(compute.DiskImageUbuntu2404)).
		WithSizeGB(50).
		WithZone("a").
		Build()

	created, err := client.Compute().Disks().Create(ctx, disk)
	if err != nil {
		fmt.Printf("   Note: Failed to create disk: %v\n", err)
		fmt.Println()
		return
	}

	start := time.Now()
	err = client.Compute().Disks().WaitForReady(ctx, *created.Metadata.Id, 2*time.Minute)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("   Failed after %v: %v\n", elapsed, err)
	} else {
		fmt.Printf("   ✓ Disk ready in %v\n", elapsed)
	}
	fmt.Println()

	// Example 3b: Fast polling for quick operations
	fmt.Println("3b. Fast polling (constant 1s interval for quick operations):")
	disk2 := compute.NewDiskBuilder("demo-disk-fast").
		WithImage(string(compute.DiskImageUbuntuMinimal2404)).
		WithSizeGB(20).
		WithZone("a").
		Build()

	created2, err := client.Compute().Disks().Create(ctx, disk2)
	if err != nil {
		fmt.Printf("   Note: Failed to create disk: %v\n", err)
		fmt.Println()
		return
	}

	start = time.Now()
	err = client.Compute().Disks().WaitForReady(
		ctx,
		*created2.Metadata.Id,
		2*time.Minute,
		compute.WithPollingInterval(1*time.Second), // Poll every 1 second
	)
	elapsed = time.Since(start)

	if err != nil {
		fmt.Printf("   Failed after %v: %v\n", elapsed, err)
	} else {
		fmt.Printf("   ✓ Disk ready in %v\n", elapsed)
	}
	fmt.Println()

	// Example 3c: Progress tracking
	fmt.Println("3c. Progress tracking (monitor wait progress):")
	disk3 := compute.NewDiskBuilder("demo-disk-progress").
		WithImage(string(compute.DiskImageRocky100)).
		WithSizeGB(30).
		WithZone("a").
		Build()

	created3, err := client.Compute().Disks().Create(ctx, disk3)
	if err != nil {
		fmt.Printf("   Note: Failed to create disk: %v\n", err)
		fmt.Println()
		return
	}

	start = time.Now()
	err = client.Compute().Disks().WaitForReady(
		ctx,
		*created3.Metadata.Id,
		2*time.Minute,
		compute.WithProgressCallback(func(attempt int, elapsed time.Duration) {
			fmt.Printf("   Attempt %d at %v...\n", attempt, elapsed.Round(time.Second))
		}),
	)
	elapsed = time.Since(start)

	if err != nil {
		fmt.Printf("   Failed after %v: %v\n", elapsed, err)
	} else {
		fmt.Printf("   ✓ Disk ready in %v\n", elapsed)
	}
	fmt.Println()

	// Example 3d: Custom exponential backoff
	fmt.Println("3d. Custom exponential backoff (aggressive polling):")
	vm := compute.NewVirtualMachineBuilder("demo-vm").
		WithBootDisk(*created.Metadata.Id).
		WithSize(string(compute.VMSizeA1aXS)).
		WithZone("a").
		Build()

	createdVM, err := client.Compute().VirtualMachines().Create(ctx, vm)
	if err != nil {
		fmt.Printf("   Note: Failed to create VM: %v\n", err)
		fmt.Println()
		return
	}

	start = time.Now()
	err = client.Compute().VirtualMachines().WaitForReady(
		ctx,
		*createdVM.Metadata.Id,
		5*time.Minute,
		compute.WithExponentialBackoff(
			1*time.Second,  // Initial interval
			15*time.Second, // Max interval
			1.5,            // Multiplier
		),
	)
	elapsed = time.Since(start)

	if err != nil {
		fmt.Printf("   Failed after %v: %v\n", elapsed, err)
	} else {
		fmt.Printf("   ✓ VM ready in %v\n", elapsed)
	}
	fmt.Println()

	// Cleanup
	fmt.Println("Cleaning up resources...")
	_ = client.Compute().VirtualMachines().Delete(ctx, *createdVM.Metadata.Id)
	_ = client.Compute().Disks().Delete(ctx, *created.Metadata.Id)
	_ = client.Compute().Disks().Delete(ctx, *created2.Metadata.Id)
	_ = client.Compute().Disks().Delete(ctx, *created3.Metadata.Id)
}

// demonstrateRetryBehavior shows how retry logic works.
func demonstrateRetryBehavior() {
	fmt.Println("--- Example 4: Automatic Retry Behavior ---")
	fmt.Println("The SDK automatically retries transient errors with exponential backoff")
	fmt.Println()

	ctx := context.Background()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Printf("Failed to create client: %v (this is expected if credentials not set)", err)
		fmt.Println()
		return
	}

	fmt.Println("Retry behavior:")
	fmt.Println("  - Retries on network errors (connection refused, timeout, etc.)")
	fmt.Println("  - Retries on 5xx server errors (500, 502, 503, 504)")
	fmt.Println("  - Retries on 429 (rate limit, when implemented)")
	fmt.Println("  - Does NOT retry on 4xx client errors (400, 404, etc.)")
	fmt.Println()
	fmt.Println("Default retry config:")
	fmt.Println("  - Max retries: 3")
	fmt.Println("  - Initial backoff: 1s")
	fmt.Println("  - Max backoff: 30s")
	fmt.Println("  - Backoff multiplier: 2x")
	fmt.Println("  - Jitter: enabled (prevents thundering herd)")
	fmt.Println()

	// Example operation - if it encounters transient errors, it will retry automatically
	fmt.Println("Attempting to list VMs (will retry automatically on transient errors)...")
	start := time.Now()
	vms, err := client.Compute().VirtualMachines().List(ctx)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("Failed after %v with retries: %v\n", elapsed, err)
	} else {
		fmt.Printf("✓ Listed %d VMs in %v (including any retry time)\n", len(vms.Items), elapsed)
	}
	fmt.Println()

	fmt.Println("Note: Retry is transparent - you don't need to implement retry logic yourself!")
	fmt.Println("The SDK handles it automatically for all operations.")
}
