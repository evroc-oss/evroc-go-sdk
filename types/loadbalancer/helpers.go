// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package loadbalancer

import "fmt"

// LoadBalancerRef is a type-safe reference to a load balancer resource.
type LoadBalancerRef string

func (r LoadBalancerRef) String() string {
	return string(r)
}

// Ref returns the fully qualified reference for use with builders and other resources.
func (lb *Loadbalancer) Ref() LoadBalancerRef {
	if lb.Metadata.Project == nil || lb.Metadata.Region == nil || lb.Metadata.Id == "" {
		return ""
	}
	return LoadBalancerRef(fmt.Sprintf("/loadbalancer/projects/%s/regions/%s/loadBalancers/%s",
		*lb.Metadata.Project, *lb.Metadata.Region, lb.Metadata.Id))
}

// Ref returns the fully qualified reference for use with builders and other resources.
func (bp *Backendpool) Ref() string {
	if bp.Metadata.Project == nil || bp.Metadata.Region == nil || bp.Metadata.Id == "" {
		return ""
	}
	return fmt.Sprintf("/loadbalancer/projects/%s/regions/%s/backendPools/%s",
		*bp.Metadata.Project, *bp.Metadata.Region, bp.Metadata.Id)
}

// Ref returns the fully qualified reference for use with builders and other resources.
func (bs *Backendservice) Ref() string {
	if bs.Metadata.Project == nil || bs.Metadata.Region == nil || bs.Metadata.Id == "" {
		return ""
	}
	return fmt.Sprintf("/loadbalancer/projects/%s/regions/%s/backendServices/%s",
		*bs.Metadata.Project, *bs.Metadata.Region, bs.Metadata.Id)
}

// Ref returns the fully qualified reference for use with builders and other resources.
func (r *L4route) Ref() string {
	if r.Metadata.Project == nil || r.Metadata.Region == nil || r.Metadata.Id == "" {
		return ""
	}
	return fmt.Sprintf("/loadbalancer/projects/%s/regions/%s/l4Routes/%s",
		*r.Metadata.Project, *r.Metadata.Region, r.Metadata.Id)
}
