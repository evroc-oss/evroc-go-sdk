// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package loadbalancer

import (
	"context"

	lbtypes "github.com/evroc-oss/evroc-go-sdk/types/loadbalancer"
)

const builderAPIVersion = "loadbalancer/" + apiVersion

// LoadBalancerBuilder provides a fluent interface for constructing LoadbalancerRequest objects.
type LoadBalancerBuilder struct {
	name      string
	ipRef     string
	listeners []lbtypes.LoadbalancerSpecListenersItem
	labels    map[string]string
}

// NewLoadBalancerBuilder creates a new builder for a LoadBalancer with the given name.
func NewLoadBalancerBuilder(name string) *LoadBalancerBuilder {
	return &LoadBalancerBuilder{name: name}
}

// WithPublicIPRef sets the public IP reference.
func (b *LoadBalancerBuilder) WithPublicIPRef(ref string) *LoadBalancerBuilder {
	b.ipRef = ref
	return b
}

// WithListener adds a listener to the load balancer.
func (b *LoadBalancerBuilder) WithListener(listener lbtypes.LoadbalancerSpecListenersItem) *LoadBalancerBuilder {
	b.listeners = append(b.listeners, listener)
	return b
}

// WithLabels sets user labels.
func (b *LoadBalancerBuilder) WithLabels(labels map[string]string) *LoadBalancerBuilder {
	b.labels = labels
	return b
}

// Build creates the LoadbalancerRequest.
func (b *LoadBalancerBuilder) Build() *lbtypes.LoadbalancerRequest {
	req := &lbtypes.LoadbalancerRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "LoadBalancer",
		Metadata:   lbtypes.RegionalMetadataRequest{Id: b.name},
		Spec: lbtypes.LoadbalancerSpec{
			PublicIPRef: b.ipRef,
		},
	}
	if len(b.listeners) > 0 {
		listeners := b.listeners
		req.Spec.Listeners = &listeners
	}
	if len(b.labels) > 0 {
		ul := lbtypes.UserLabels(b.labels)
		req.Metadata.UserLabels = &ul
	}
	return req
}

// Create builds and creates the LoadBalancer in one call.
func (b *LoadBalancerBuilder) Create(ctx context.Context, client *LoadBalancersService) (*lbtypes.Loadbalancer, error) {
	return client.Create(ctx, b.Build())
}

// BackendPoolBuilder provides a fluent interface for constructing BackendpoolRequest objects.
type BackendPoolBuilder struct {
	name        string
	backendRefs []string
	labels      map[string]string
}

// NewBackendPoolBuilder creates a new builder for a BackendPool with the given name.
func NewBackendPoolBuilder(name string) *BackendPoolBuilder {
	return &BackendPoolBuilder{name: name}
}

// WithBackendRefs sets the backend VM references.
func (b *BackendPoolBuilder) WithBackendRefs(refs []string) *BackendPoolBuilder {
	b.backendRefs = refs
	return b
}

// WithLabels sets user labels.
func (b *BackendPoolBuilder) WithLabels(labels map[string]string) *BackendPoolBuilder {
	b.labels = labels
	return b
}

// Build creates the BackendpoolRequest.
func (b *BackendPoolBuilder) Build() *lbtypes.BackendpoolRequest {
	req := &lbtypes.BackendpoolRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "BackendPool",
		Metadata:   lbtypes.RegionalMetadataRequest{Id: b.name},
	}
	if len(b.backendRefs) > 0 {
		refs := b.backendRefs
		req.Spec.BackendRefs = &refs
	}
	if len(b.labels) > 0 {
		ul := lbtypes.UserLabels(b.labels)
		req.Metadata.UserLabels = &ul
	}
	return req
}

// Create builds and creates the BackendPool in one call.
func (b *BackendPoolBuilder) Create(ctx context.Context, client *BackendPoolsService) (*lbtypes.Backendpool, error) {
	return client.Create(ctx, b.Build())
}

// BackendServiceBuilder provides a fluent interface for constructing BackendserviceRequest objects.
type BackendServiceBuilder struct {
	name           string
	port           int32
	backendPoolRef string
	proxyProtocol  bool
	healthCheck    *lbtypes.BackendserviceSpecHealthCheck
	labels         map[string]string
}

// NewBackendServiceBuilder creates a new builder for a BackendService with the given name.
func NewBackendServiceBuilder(name string) *BackendServiceBuilder {
	return &BackendServiceBuilder{name: name}
}

// WithPort sets the backend port.
func (b *BackendServiceBuilder) WithPort(port int32) *BackendServiceBuilder {
	b.port = port
	return b
}

// WithBackendPoolRef sets the backend pool reference.
func (b *BackendServiceBuilder) WithBackendPoolRef(ref string) *BackendServiceBuilder {
	b.backendPoolRef = ref
	return b
}

// WithProxyProtocol enables PROXY protocol.
func (b *BackendServiceBuilder) WithProxyProtocol(enabled bool) *BackendServiceBuilder {
	b.proxyProtocol = enabled
	return b
}

// WithHealthCheck sets a custom health check configuration.
func (b *BackendServiceBuilder) WithHealthCheck(hc *lbtypes.BackendserviceSpecHealthCheck) *BackendServiceBuilder {
	b.healthCheck = hc
	return b
}

// WithTCPHealthCheck configures a TCP health check on the backend service port.
func (b *BackendServiceBuilder) WithTCPHealthCheck() *BackendServiceBuilder {
	b.healthCheck = &lbtypes.BackendserviceSpecHealthCheck{
		Tcp: &struct {
			Receive *string `json:"receive,omitempty"`
			Send    *string `json:"send,omitempty"`
		}{},
	}
	return b
}

// WithHTTPHealthCheck configures an HTTP health check on the given path.
func (b *BackendServiceBuilder) WithHTTPHealthCheck(path string) *BackendServiceBuilder {
	b.healthCheck = &lbtypes.BackendserviceSpecHealthCheck{
		Http: &struct {
			ExpectedStatuses *[]int32                                    `json:"expectedStatuses,omitempty"`
			Host             *string                                     `json:"host,omitempty"`
			Method           *lbtypes.BackendserviceSpecHealthCheckHttpMethod `json:"method,omitempty"`
			Path             string                                      `json:"path"`
		}{
			Path: path,
		},
	}
	return b
}

// WithLabels sets user labels.
func (b *BackendServiceBuilder) WithLabels(labels map[string]string) *BackendServiceBuilder {
	b.labels = labels
	return b
}

// Build creates the BackendserviceRequest.
func (b *BackendServiceBuilder) Build() *lbtypes.BackendserviceRequest {
	req := &lbtypes.BackendserviceRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "BackendService",
		Metadata:   lbtypes.RegionalMetadataRequest{Id: b.name},
		Spec: lbtypes.BackendserviceSpec{
			Port: b.port,
		},
	}
	if b.backendPoolRef != "" {
		req.Spec.BackendPoolRef = &b.backendPoolRef
	}
	if b.proxyProtocol {
		pp := true
		req.Spec.ProxyProtocol = &pp
	}
	if b.healthCheck != nil {
		req.Spec.HealthCheck = b.healthCheck
	}
	if len(b.labels) > 0 {
		ul := lbtypes.UserLabels(b.labels)
		req.Metadata.UserLabels = &ul
	}
	return req
}

// Create builds and creates the BackendService in one call.
func (b *BackendServiceBuilder) Create(ctx context.Context, client *BackendServicesService) (*lbtypes.Backendservice, error) {
	return client.Create(ctx, b.Build())
}

// L4RouteBuilder provides a fluent interface for constructing L4routeRequest objects.
type L4RouteBuilder struct {
	name                  string
	backendServiceRef     string
	labels                map[string]string
}

// NewL4RouteBuilder creates a new builder for an L4Route with the given name.
func NewL4RouteBuilder(name string) *L4RouteBuilder {
	return &L4RouteBuilder{name: name}
}

// WithBackendServiceRef sets the default backend service reference.
func (b *L4RouteBuilder) WithBackendServiceRef(ref string) *L4RouteBuilder {
	b.backendServiceRef = ref
	return b
}

// WithLabels sets user labels.
func (b *L4RouteBuilder) WithLabels(labels map[string]string) *L4RouteBuilder {
	b.labels = labels
	return b
}

// Build creates the L4routeRequest.
func (b *L4RouteBuilder) Build() *lbtypes.L4routeRequest {
	req := &lbtypes.L4routeRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "L4Route",
		Metadata:   lbtypes.RegionalMetadataRequest{Id: b.name},
		Spec: lbtypes.L4routeSpec{
			DefaultBackendServiceRef: b.backendServiceRef,
		},
	}
	if len(b.labels) > 0 {
		ul := lbtypes.UserLabels(b.labels)
		req.Metadata.UserLabels = &ul
	}
	return req
}

// Create builds and creates the L4Route in one call.
func (b *L4RouteBuilder) Create(ctx context.Context, client *L4RoutesService) (*lbtypes.L4route, error) {
	return client.Create(ctx, b.Build())
}
