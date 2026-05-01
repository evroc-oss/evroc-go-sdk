// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Authentication Example
//
// Demonstrates different ways to authenticate with the evroc Go SDK.
//
// Method 1: Using evroc CLI (Recommended for development)
//
//	evroc login
//	go run main.go
//
// Method 2: Using environment variables
//
//	export EVROC_REFRESH_TOKEN="your-refresh-token"
//	export EVROC_PROJECT="project-uuid"
//	export EVROC_REGION="se-sto"
//	go run main.go
//
// Method 3: Using explicit file path
//
//	# Load from default CLI config location
//	go run main.go -method file
//
//	# Or from custom path
//	go run main.go -method file -config /path/to/.evroc/config.yaml
//
// The example tests authentication by listing virtual machines in your project.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	evroc "github.com/evroc-oss/evroc-go-sdk"
)

var (
	method     = flag.String("method", "auto", "Authentication method: auto, cli, file")
	configPath = flag.String("config", "", "Path to config file")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	var client *evroc.Client
	var err error

	switch *method {
	case "auto":
		client, err = evroc.NewFromEnv(ctx)
	case "cli":
		client, err = evroc.NewFromCLIConfig(ctx, "")
	case "file":
		path := *configPath
		if path == "" {
			path = os.ExpandEnv("$HOME/.evroc/config.yaml")
		}
		client, err = evroc.NewFromCLIConfig(ctx, path)
	default:
		log.Fatalf("Unknown method: %s", *method)
	}

	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Printf("Authenticated successfully\n")
	fmt.Printf("Project: %s, Region: %s\n", client.DefaultProject(), client.DefaultRegion())

	vms, err := client.Compute().VirtualMachines().List(ctx)
	if err != nil {
		log.Fatalf("Failed to list VMs: %v", err)
	}

	fmt.Printf("Found %d VM(s)\n", len(vms.Items))
}
