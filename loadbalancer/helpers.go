// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package loadbalancer

import (
	"github.com/evroc-oss/evroc-go-sdk/metrics"
	lbtypes "github.com/evroc-oss/evroc-go-sdk/types/loadbalancer"
)

// WithMetrics enables metrics collection for this client.
func (c *Client) WithMetrics(m *metrics.Manager) *Client {
	c.metrics = m
	return c
}

// IsReady returns true if the LoadBalancer has a Ready=True condition.
func IsReady(lb *lbtypes.Loadbalancer) bool {
	if lb == nil || lb.Status.Conditions == nil {
		return false
	}
	for _, cond := range *lb.Status.Conditions {
		if cond.Type == "Ready" && string(cond.Status) == "True" {
			return true
		}
	}
	return false
}
