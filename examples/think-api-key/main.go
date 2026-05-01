// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Think API key example
//
// Demonstrates operations with a shared model API key (which is distinct
// from your evroc login token, for compatibility with the usual long-lived
// token model for AI integrations).
//
// Since we only call the OpenAI-compatible model listing endpoint, this
// example will consume no tokens and should not result in any billing.
//
// # Running
//
//	export EVROC_TOKEN=<token>
//	export EVROC_PROJECT=<project-id>
//
//	go run main.go
//
// # What This Example Shows
//
//  1. Creating an API key
//  2. Making a /v1/models request with the API key
//  3. Deleting the API key
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/think"
)

// this is the model serving endpoint, which is authenticated using a
// shared models API key instead of your evroc token
const sharedModelsV1ModelsURL = "https://models.think.evroc.com/v1/models"

// these structs are a minimal version of the OpenAI-compatible /v1/models spec
type OpenAIV1ModelsItem struct {
	Id string `json:"id"`
}

type OpenAIV1Models struct {
	Data []OpenAIV1ModelsItem `json:"data"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatal(err)
	}

	keyId := "think-example-" + strings.ToLower(rand.Text())[:6]

	// 1. Creating key
	fmt.Println("1. Creating key...")
	result, err := think.NewAPIKeyBuilder(keyId).WithExpiryTimestamp(time.Now().Add(time.Hour)).Create(ctx, client.Think().ApiKeys())
	if err != nil {
		log.Fatalf("Error creating API key: %v", err)
	}
	fmt.Printf("   ✓ Created key: %s\n", keyId)
	if result.Status.Token == nil {
		log.Fatalf("No api token received in response: %v", result)
	}
	fmt.Printf("   ✓ Received token: %s\n", *result.Status.Token)

	// 2. Call shared models API
	fmt.Println("\n2. Calling shared models service using API key...")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sharedModelsV1ModelsURL, bytes.NewReader([]byte{}))
	if err != nil {
		log.Fatalf("Error constructing shared models request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+*result.Status.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error calling shared models service: %v", err)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	var body OpenAIV1Models
	if err = json.Unmarshal(bodyBytes, &body); err != nil {
		log.Fatalf("Error parsing response body: %v", err)
	}
	fmt.Printf("   ✓ Read %s\n", sharedModelsV1ModelsURL)
	for _, item := range body.Data {
		fmt.Println("- " + item.Id)
	}

	// 3. Delete API key
	fmt.Println("\n3. Deleting API key...")
	err = client.Think().ApiKeys().Delete(ctx, keyId)
	if err != nil {
		log.Fatalf("Error deleting API key: %v", err)
	}
	fmt.Printf("   ✓ Deleted\n")
}
