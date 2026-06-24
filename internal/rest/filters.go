// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"net/url"

	"github.com/evroc-oss/evroc-go-sdk/filter"
)

// ListFilter is an alias for the public filter.ListFilter interface.
// Generated service code references rest.ListFilter; this alias ensures
// that external consumers can pass filter.WithLabelSelector() values
// directly to List() methods.
type ListFilter = filter.ListFilter

// WithLabelSelector delegates to filter.WithLabelSelector.
var WithLabelSelector = filter.WithLabelSelector

// WithLabels delegates to filter.WithLabels.
var WithLabels = filter.WithLabels

// ListWithFilters applies filters and calls ListResourcesWithQuery.
// Used by generated List() methods to support composable filters.
func ListWithFilters[T any](ctx context.Context, client *Client, path string, filters ...ListFilter) (T, error) {
	query := url.Values{}
	for _, f := range filters {
		f.Apply(query)
	}
	return ListResourcesWithQuery[T](ctx, client, path, query)
}
