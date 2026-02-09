// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package iam

import (
	"github.com/evroc-oss/evroc-go-sdk/metrics"
)

// ============================================================================
// Metrics Support
// ============================================================================

// WithMetrics enables metrics collection for this iam client.
// Returns the client to allow chaining.
func (c *Client) WithMetrics(m *metrics.Manager) *Client {
	c.metrics = m
	return c
}
