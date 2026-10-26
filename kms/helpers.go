// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package kms

import "github.com/evroc-oss/evroc-go-sdk/metrics"

// WithMetrics enables metrics collection for this client.
func (c *Client) WithMetrics(m *metrics.Manager) *Client {
	c.metrics = m
	return c
}
