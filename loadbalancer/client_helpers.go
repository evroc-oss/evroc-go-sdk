// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package loadbalancer

import "fmt"

// LoadBalancerRef creates a fully-qualified load balancer reference from a name.
func (c *Client) LoadBalancerRef(name string) string {
	return fmt.Sprintf("/loadbalancer/projects/%s/regions/%s/loadBalancers/%s",
		c.parent.DefaultProject(), c.parent.DefaultRegion(), name)
}

// BackendPoolRef creates a fully-qualified backend pool reference from a name.
func (c *Client) BackendPoolRef(name string) string {
	return fmt.Sprintf("/loadbalancer/projects/%s/regions/%s/backendPools/%s",
		c.parent.DefaultProject(), c.parent.DefaultRegion(), name)
}

// BackendServiceRef creates a fully-qualified backend service reference from a name.
func (c *Client) BackendServiceRef(name string) string {
	return fmt.Sprintf("/loadbalancer/projects/%s/regions/%s/backendServices/%s",
		c.parent.DefaultProject(), c.parent.DefaultRegion(), name)
}

// L4RouteRef creates a fully-qualified L4 route reference from a name.
func (c *Client) L4RouteRef(name string) string {
	return fmt.Sprintf("/loadbalancer/projects/%s/regions/%s/l4Routes/%s",
		c.parent.DefaultProject(), c.parent.DefaultRegion(), name)
}
