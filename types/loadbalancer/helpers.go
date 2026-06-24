// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package loadbalancer

import "fmt"

// LoadBalancerRef is a type-safe reference to a load balancer resource.
type LoadBalancerRef string

func (r LoadBalancerRef) String() string {
	return string(r)
}

func (lb *Loadbalancer) Ref() LoadBalancerRef {
	if lb.Metadata.Project == nil || lb.Metadata.Region == nil || lb.Metadata.Id == "" {
		return ""
	}
	return LoadBalancerRef(fmt.Sprintf("/loadbalancer/projects/%s/regions/%s/loadBalancers/%s",
		*lb.Metadata.Project, *lb.Metadata.Region, lb.Metadata.Id))
}
