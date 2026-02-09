// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package main demonstrates SDK metrics integration with Prometheus.
// This example continuously creates and deletes disks to generate metrics.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/compute"
	"github.com/evroc-oss/evroc-go-sdk/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	diskNamePrefix = "metrics-demo-disk"
	diskImage      = "ubuntu-minimal.24-04.1"
	diskSizeGB     = 10
	zone           = "a"
	cycleInterval  = 30 * time.Second
)

func main() {
	// Create context that cancels on SIGINT/SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Println("=== SDK Metrics Example ===")
	fmt.Println("\nThis example continuously creates and deletes disks to generate Prometheus metrics.")
	fmt.Println("Metrics are exposed at http://localhost:9090/metrics")
	fmt.Println("\nPress Ctrl+C to stop\n")

	// Step 1: Create metrics manager
	metricsManager := metrics.NewManager()
	fmt.Println("Metrics manager created")

	// Step 2: Start metrics HTTP server
	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(metricsManager.Registry(), promhttp.HandlerOpts{}))
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()
	fmt.Println("Metrics server started at http://localhost:9090/metrics")

	// Step 3: Create SDK client with metrics
	client, err := evroc.NewFromEnv(ctx, evroc.WithMetrics(metricsManager))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("SDK client created with metrics instrumentation")

	// Step 4: Print example metrics info
	fmt.Println("\nAvailable Metrics:")
	fmt.Println("  - evroc_sdk_api_calls_total - Total API calls")
	fmt.Println("  - evroc_sdk_api_calls_duration_seconds - API call latency")
	fmt.Println("  - evroc_sdk_api_calls_errors_total - API errors")
	fmt.Println("  - evroc_sdk_retries_total - Retry attempts")
	fmt.Println("  - evroc_sdk_waiter_operations_total - Waiter operations")
	fmt.Println("  - evroc_sdk_waiter_duration_seconds - Waiter duration")
	fmt.Println("  - evroc_sdk_auth_token_refreshes_total - Token refreshes")

	fmt.Println("\nStarting disk create/delete cycle...")
	fmt.Printf("Creating and deleting disk every %v\n\n", cycleInterval)

	// Step 5: Run create/delete cycle
	cycle := 0
	ticker := time.NewTicker(cycleInterval)
	defer ticker.Stop()

	// Run first cycle immediately
	runCycle(ctx, client, cycle)
	cycle++

	// Then run on ticker
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n\nShutting down gracefully...")
			return
		case <-ticker.C:
			runCycle(ctx, client, cycle)
			cycle++
		}
	}
}

func runCycle(ctx context.Context, client *evroc.Client, cycle int) {
	diskName := fmt.Sprintf("%s-%d", diskNamePrefix, cycle)

	fmt.Printf("[Cycle %d] Creating disk '%s'...\n", cycle, diskName)

	// Create disk - generates API call metrics
	diskReq := compute.NewDiskBuilder(diskName).
		WithImage(diskImage).
		WithSizeGB(diskSizeGB).
		WithZone(zone).
		Build()

	disk, err := client.Compute().Disks().Create(ctx, diskReq)
	if err != nil {
		log.Printf("[Cycle %d] Failed to create disk: %v\n", cycle, err)
		return
	}
	fmt.Printf("[Cycle %d] Disk created: %s\n", cycle, *disk.Metadata.Id)

	// Wait for disk to be ready - generates waiter metrics
	fmt.Printf("[Cycle %d] Waiting for disk to be ready...\n", cycle)
	err = client.Compute().Disks().WaitForReady(ctx, diskName, 2*time.Minute)
	if err != nil {
		log.Printf("[Cycle %d] Wait timeout: %v\n", cycle, err)
		// Continue to delete even if wait fails
	} else {
		fmt.Printf("[Cycle %d] Disk is ready\n", cycle)
	}

	// Delete disk - generates API call metrics
	fmt.Printf("[Cycle %d] Deleting disk '%s'...\n", cycle, diskName)
	err = client.Compute().Disks().Delete(ctx, diskName)
	if err != nil {
		log.Printf("[Cycle %d] Failed to delete disk: %v\n", cycle, err)
		return
	}
	fmt.Printf("[Cycle %d] Disk deleted\n", cycle)

	// Wait for deletion - generates waiter metrics
	fmt.Printf("[Cycle %d] Waiting for disk deletion...\n", cycle)
	err = client.Compute().Disks().WaitForDeleted(ctx, diskName, 2*time.Minute)
	if err != nil {
		log.Printf("[Cycle %d] Delete wait timeout: %v\n", cycle, err)
	} else {
		fmt.Printf("[Cycle %d] Disk fully deleted\n", cycle)
	}

	fmt.Printf("[Cycle %d] Cycle complete\n\n", cycle)
}
