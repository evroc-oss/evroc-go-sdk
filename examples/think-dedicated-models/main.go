// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Think dedicated models example
//
// Demonstrates listing, creating, waiting for, using and finally destroying
// a dedicated model instance. Note that this will create a real dedicated
// model instance with associated billing.
//
// # Running
//
//	export EVROC_TOKEN=<token>
//	export EVROC_PROJECT=<project-id>
//	export EVROC_REGION=se-sto
//
//	go run main.go
//
// # What This Example Shows
//
//  1. Fetching the desired model
//  2. Listing available instance sizes
//  3. Creating a dedicated instance (stopped)
//  4. Start the instance
//  5. Waiting for the dedicated instance to be ready
//  6. Calling the serving endpoint
//  7. Stop the instance
//  8. Waiting for the instance to be stopped
//  9. Deleting the dedicated instance

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/think"
)

const (
	instanceName = "think-example-instance"
	modelName    = "openai-gpt-oss-120b"
	sizeName     = "1-b200-27c-240g"

	// Timeouts.
	instanceReadyTimeout = 5 * time.Minute
)

func main() {
	// Create context that cancels on SIGINT/SIGTERM (Ctrl-C)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create client - config has project/region
	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// 1. Fetch the model
	fmt.Println("1. Fetching model spec")
	model, err := client.Think().Models().Get(ctx, modelName)
	if err != nil {
		log.Fatalf("Failed to fetch model spec: %v", err)
	}
	fmt.Printf("   ✓ Found %s\n", model.Metadata.Id)

	// 2. Fetch the size
	fmt.Println("\n2. Fetching instance size specs")
	sizes, err := client.Think().Sizes().List(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch sizes: %v", err)
	}
	found := false
	for _, size := range sizes.Items {
		if size.Metadata.Id == sizeName {
			found = true
		}
	}
	if !found {
		log.Fatalf("Expected size %s was not found", sizeName)
	}
	fmt.Printf("   ✓ Found %s\n", sizeName)

	// 3. Create the instance
	fmt.Println("\n3. Creating instance with builder pattern...")
	instance, err := think.NewInstanceBuilder(instanceName).
		WithModel(modelName).
		WithSize(sizeName).
		WithStopped(true).
		Create(ctx, client.Think().Instances())
	if err != nil {
		log.Fatalf("Failed to create instance: %v", err)
	}
	fmt.Printf("   ✓ Created instance: %s\n", instance.Metadata.Id)
	fmt.Printf("   ✓ Instance endpoint: %s\n", *instance.Status.Endpoint)

	// 4. Start the instance
	// Note that this step could be skipped if the instance was created without
	// WithStopped(true); this just demonstrates start/stop of an existing instance
	fmt.Println("\n3. Starting the instance...")
	err = client.Think().Instances().Start(ctx, instance.Metadata.Id)
	if err != nil {
		log.Fatalf("Failed to start instance: %v", err)
	}
	fmt.Printf("   ✓ Instance start requested\n")

	// 5. Wait for the instance to be ready
	fmt.Println("\n5. Waiting for instance to be ready...")
	if _, err := client.Think().Instances().WaitForReady(ctx, instance.Metadata.Id, instanceReadyTimeout); err != nil {
		log.Fatalf("Error waiting for instance to be ready: %v", err)
	}
	fmt.Printf("   ✓ Instance is ready\n")

	// 6. Call the instance
	fmt.Println("\n6. Calling /v1/chat/completions on the instance...")
	url := fmt.Sprintf("%s/v1/chat/completions", *instance.Status.Endpoint)
	body := map[string]any{
		"model": *model.Spec.Handle,
		"messages": []map[string]string{
			{"role": "user", "content": "how about a nice game of chess?"},
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		log.Fatalf("Error marshaling request body: %v", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	req.Header.Add("Authorization", "Bearer "+*instance.Status.Token)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		log.Fatalf("Error constructing request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error calling model: %v", err)
	}
	fmt.Printf("   ✓ Instance responded\n")
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Printf("%s: %s\n", modelName, string(bodyBytes))

	// 7. Stop the instance
	// Note that this step could be skipped; you can delete a running instance,
	// but this demonstrates stopping an instance which retains the config but
	// removes the running resources
	fmt.Println("\n7. Stopping the instance...")
	err = client.Think().Instances().Stop(ctx, instance.Metadata.Id)
	if err != nil {
		log.Fatalf("Failed to stop instance: %v", err)
	}
	fmt.Printf("   ✓ Instance stop requested\n")

	// 8. Wait for the instance to be stopped
	fmt.Println("\n8. Waiting for instance to be stopped...")
	if _, err := client.Think().Instances().WaitForStopped(ctx, instance.Metadata.Id, instanceReadyTimeout); err != nil {
		log.Fatalf("Error waiting for instance to be stopped: %v", err)
	}
	fmt.Printf("   ✓ Instance is stopped\n")

	// 9. Delete the instance
	fmt.Println("\n9. Deleting the instance...")
	err = client.Think().Instances().Delete(ctx, instance.Metadata.Id)
	if err != nil {
		log.Fatalf("Error deleting instance: %v", err)
	}
	fmt.Printf("   ✓ Instance deleted\n")
}
