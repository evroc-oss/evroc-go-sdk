// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package filter provides composable filters for List() methods.
package filter

import (
	"fmt"
	"net/url"
)

const queryParamLabelSelector = "labelSelector"

// ListFilter defines a composable filter for List() methods.
//
// Filters are passed as variadic arguments to any service's List() method:
//
//	disks, err := client.Compute().Disks().List(ctx,
//		filter.WithLabelSelector(map[string]string{"env": "prod"}),
//	)
type ListFilter interface {
	Apply(url.Values)
}

// filterFunc adapts functions to ListFilter.
type filterFunc func(url.Values)

func (f filterFunc) Apply(v url.Values) {
	f(v)
}

// WithLabelSelector creates a filter from a Kubernetes-style label selector string.
// This is the primary API for CLI usage — it passes the selector string through
// to the API verbatim, supporting the full selector syntax:
//
//	filter.WithLabelSelector("env=prod")
//	filter.WithLabelSelector("env=prod,team=platform")
//	filter.WithLabelSelector("team in (frontend,backend)")
//	filter.WithLabelSelector("networking.evroc.com/managed-network=default")
func WithLabelSelector(selector string) ListFilter {
	return filterFunc(func(v url.Values) {
		if selector != "" {
			v.Set(queryParamLabelSelector, selector)
		}
	})
}

// WithLabels creates a filter from a map of key=value label pairs.
// For simple equality matching only. For advanced selectors (in, notin, exists),
// use WithLabelSelector with a raw string.
//
//	filter.WithLabels(map[string]string{"env": "prod", "team": "platform"})
func WithLabels(labels map[string]string) ListFilter {
	return filterFunc(func(v url.Values) {
		if len(labels) > 0 {
			v.Set(queryParamLabelSelector, formatLabelsForQuery(labels))
		}
	})
}

func formatLabelsForQuery(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	result := ""
	for k, v := range labels {
		if result != "" {
			result += ","
		}
		result += fmt.Sprintf("%s=%s", k, v)
	}
	return result
}
