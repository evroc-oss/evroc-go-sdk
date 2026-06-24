// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package loadbalancer provides access to the evroc LoadBalancer v1alpha1 API.
//
// # Resources
//
//   - LoadBalancers: L4 (TCP) load balancers with listeners
//   - BackendPools: Groups of VM endpoints
//   - BackendServices: Port and proxy-protocol configuration per listener
//   - L4Routes: Layer 4 routing rules linking listeners to backend services
//
// # Usage
//
// Use the builders to construct resource requests:
//
//	lbs := client.LoadBalancer()
//
//	pool, _ := loadbalancer.NewBackendPoolBuilder("my-pool").
//	    WithBackendRefs([]string{lbs.BackendPoolRef("vm-1")}).
//	    Create(ctx, lbs.BackendPools())
//
//	svc, _ := loadbalancer.NewBackendServiceBuilder("my-svc").
//	    WithPort(8080).
//	    WithBackendPoolRef(lbs.BackendPoolRef("my-pool")).
//	    Create(ctx, lbs.BackendServices())
//
//	route, _ := loadbalancer.NewL4RouteBuilder("my-route").
//	    WithBackendServiceRef(lbs.BackendServiceRef("my-svc")).
//	    Create(ctx, lbs.L4Routes())
//
//	lb, _ := loadbalancer.NewLoadBalancerBuilder("my-lb").
//	    WithListener(lbtypes.LoadbalancerSpecListenersItem{...}).
//	    Create(ctx, lbs.LoadBalancers())
//
// Or use the service APIs directly for full control:
//
//	lbs.BackendPools().Create(ctx, req)
//	lbs.BackendPools().Get(ctx, name)
//	lbs.BackendPools().Patch(ctx, name, patch)
//	lbs.BackendPools().Delete(ctx, name)
package loadbalancer
