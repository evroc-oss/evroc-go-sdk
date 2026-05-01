// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package think provides access to the evroc Think API.
//
// The Think API enables management of AI model inference resources:
//   - API Keys: Create and manage API keys for model access
//   - Models: List available models for dedicated instances
//   - Shared Models: Discover shared models callable via API key
//   - Sizes: List available GPU instance sizes
//   - Instances: Full lifecycle management of dedicated model-serving GPU instances
//
// # Getting Started
//
// List available models:
//
//	client, err := evroc.NewFromEnv(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	models, err := client.Think().Models().List(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Instance Lifecycle
//
// Create and manage dedicated GPU instances:
//
//	instance, err := client.Think().Instances().Create(ctx, request)
//	// ... wait for instance to be ready ...
//	err = client.Think().Instances().Start(ctx, "my-instance")
//	err = client.Think().Instances().Stop(ctx, "my-instance")
package think
