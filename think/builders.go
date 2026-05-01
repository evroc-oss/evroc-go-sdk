// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package think

import (
	"context"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/types/think"
)

const builderAPIVersion = "think/" + apiVersion

type APIKeyBuilder struct {
	id              string
	expiryTimestamp *time.Time
}

type InstanceBuilder struct {
	id      string
	model   string
	size    *string
	stopped *bool
}

// NewAPIKeyBuilder creates an APIKeyBuilder with default configuration
func NewAPIKeyBuilder(id string) *APIKeyBuilder {
	return &APIKeyBuilder{id: id}
}

// WithExpiryTimestamp sets the expiry timestamp of the key
func (b *APIKeyBuilder) WithExpiryTimestamp(expiryTimestamp time.Time) *APIKeyBuilder {
	b.expiryTimestamp = &expiryTimestamp
	return b
}

// Build constructs the request to create an API key
func (b *APIKeyBuilder) Build() *think.ApikeyRequest {
	return &think.ApikeyRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "APIKey",
		Metadata:   think.GlobalProjectMetadataRequest{Id: b.id},
		Spec:       think.ApikeySpec{ExpiryTimestamp: b.expiryTimestamp},
	}
}

// Create builds the request and executes it, returning the new API key
func (b *APIKeyBuilder) Create(ctx context.Context, client *ApiKeysService) (*think.Apikey, error) {
	apiKeyReq := b.Build()
	return client.Create(ctx, apiKeyReq)
}

// NewInstanceBuilder creates an InstanceBuilder
func NewInstanceBuilder(id string) *InstanceBuilder {
	return &InstanceBuilder{id: id}
}

// WithModel sets the model (the Id of a Model resource) for this Instance
func (b *InstanceBuilder) WithModel(model string) *InstanceBuilder {
	b.model = model
	return b
}

// WithSize sets the size (the Id of a Size resource) for this Instance
func (b *InstanceBuilder) WithSize(size string) *InstanceBuilder {
	b.size = &size
	return b
}

// WithStopped sets whether the instance should be created in a stopped state
// with no resources allocated
func (b *InstanceBuilder) WithStopped(stopped bool) *InstanceBuilder {
	b.stopped = &stopped
	return b
}

// Build constructs the request to create an Instance
func (b *InstanceBuilder) Build() *think.InstanceRequest {
	return &think.InstanceRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "Instance",
		Metadata:   think.RegionalMetadataRequest{Id: b.id},
		Spec: think.InstanceSpec{
			Model:   b.model,
			Size:    b.size,
			Stopped: b.stopped,
		},
	}
}

// Create builds the request and executes it, returning the new Instnace
func (b *InstanceBuilder) Create(ctx context.Context, client *InstancesService) (*think.Instance, error) {
	instanceReq := b.Build()
	return client.Create(ctx, instanceReq)
}
