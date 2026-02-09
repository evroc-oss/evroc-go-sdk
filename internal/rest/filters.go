// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package rest

import (
	"context"
	"fmt"
	"net/url"
)

const (
	// QueryParamLabelSelector is the query parameter for label filtering.
	QueryParamLabelSelector = "labelSelector"
)

// ListFilter defines a composable filter for List() methods.
type ListFilter interface {
	Apply(url.Values)
}

// filterFunc adapts functions to ListFilter.
type filterFunc func(url.Values)

// Apply implements ListFilter.
func (f filterFunc) Apply(v url.Values) {
	f(v)
}

// WithLabelSelector creates a label filter.
// User labels: {"env": "prod"}
// System labels: {"compute.evroc.com/zone": "a"}
func WithLabelSelector(labels map[string]string) ListFilter {
	return filterFunc(func(v url.Values) {
		if len(labels) > 0 {
			v.Set(QueryParamLabelSelector, formatLabelsForQuery(labels))
		}
	})
}

// formatLabelsForQuery converts labels to query format (key1=value1,key2=value2).
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

// ListWithFilters applies filters and calls ListResourcesWithQuery.
// Used by generated List() methods to support composable filters.
func ListWithFilters[T any](ctx context.Context, client *Client, path string, filters ...ListFilter) (T, error) {
	query := url.Values{}
	for _, filter := range filters {
		filter.Apply(query)
	}
	return ListResourcesWithQuery[T](ctx, client, path, query)
}
