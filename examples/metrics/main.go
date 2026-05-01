// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

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
	cycleInterval  = 5 * time.Second // Quick cycles to generate metrics
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
	fmt.Println("\nExample Metrics:")
	fmt.Println("  - evroc_sdk_api_calls_total{method=\"POST\",path=\"/compute/.../disks\",status=\"201\"}")
	fmt.Println("  - evroc_sdk_api_calls_duration_seconds{method=\"DELETE\",path=\"/compute/.../disks/{id}\"}")
	fmt.Println("  - evroc_sdk_api_calls_errors_total{method=\"...\",path=\"...\"}")
	fmt.Println("\nTry: curl http://localhost:9090/metrics | grep evroc_sdk")

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
	fmt.Printf("[Cycle %d] Disk created: %s\n", cycle, disk.Metadata.Id)

	// Short delay to allow disk creation to begin
	time.Sleep(2 * time.Second)

	// Delete disk immediately - generates API call metrics
	// We don't wait for ready since this is just a metrics demo
	fmt.Printf("[Cycle %d] Deleting disk '%s'...\n", cycle, diskName)
	err = client.Compute().Disks().Delete(ctx, diskName)
	if err != nil {
		log.Printf("[Cycle %d] Failed to delete disk: %v\n", cycle, err)
		return
	}
	fmt.Printf("[Cycle %d] Disk deleted\n", cycle)

	fmt.Printf("[Cycle %d] Cycle complete\n\n", cycle)
}
