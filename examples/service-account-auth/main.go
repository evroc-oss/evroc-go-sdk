// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Service Account Authentication Example
//
// Demonstrates how to authenticate with a service account credential.
// This is the recommended authentication method for automated workloads (CI/CD,
// Terraform, long-running services) that don't have an interactive user.
//
// Prerequisites:
//
//  1. Create a service account and credential via the IAM API.
//     The API response includes the secret (a private key as JWK).
//
//  2. Set the required environment variables (see below).
//
// EVROC_SERVICE_ACCOUNT_SECRET accepts either a file path or a base64-encoded string.
// Using base64 is convenient in CI/CD where the key is injected as a secret.
//
// Usage:
//
//	# From a file:
//	export EVROC_SERVICE_ACCOUNT_ID=myserviceaccount
//	export EVROC_SERVICE_ACCOUNT_SECRET=./private.jwk
//	export EVROC_PROJECT=my-project-1234
//	export EVROC_REGION=se-sto
//	go run main.go
//
//	# Or with base64 (e.g. from a CI secret):
//	export EVROC_SERVICE_ACCOUNT_SECRET=$(base64 < private.jwk)
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	evroc "github.com/evroc-oss/evroc-go-sdk"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Println("Authenticated successfully with service account")
	fmt.Printf("Project: %s, Region: %s\n", client.DefaultProject(), client.DefaultRegion())

	// Verify by listing VMs
	vms, err := client.Compute().VirtualMachines().List(ctx)
	if err != nil {
		log.Fatalf("API call failed: %v", err)
	}
	fmt.Printf("Found %d VM(s)\n", len(vms.Items))
}
