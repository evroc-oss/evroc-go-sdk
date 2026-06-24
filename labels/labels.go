// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package labels provides generic label management for any SDK service.
package labels

import "context"

// Service defines the operations needed for label management on any resource.
type Service[T any] interface {
	Get(ctx context.Context, name string) (T, error)
	Patch(ctx context.Context, name string, patch interface{}) (T, error)
}

// ExtractFunc extracts user labels from a resource.
type ExtractFunc[T any] func(T) map[string]string

// Helper provides label CRUD operations for a specific resource type.
// Construct one via For().
//
//	h := labels.For(diskService, func(d *compute.Disk) map[string]string {
//		if d.Metadata.UserLabels == nil { return nil }
//		return map[string]string(*d.Metadata.UserLabels)
//	})
//	current, _ := h.Get(ctx, "my-disk")
//	h.Add(ctx, "my-disk", map[string]string{"env": "prod"})
type Helper[T any] struct {
	svc     Service[T]
	extract ExtractFunc[T]
}

// For creates a label helper for the given service and label extractor.
func For[T any](svc Service[T], extract ExtractFunc[T]) *Helper[T] {
	return &Helper[T]{svc: svc, extract: extract}
}

// Get returns the current user labels for the named resource.
func (h *Helper[T]) Get(ctx context.Context, name string) (map[string]string, error) {
	resource, err := h.svc.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	result := h.extract(resource)
	if result == nil {
		return map[string]string{}, nil
	}
	return result, nil
}

// Set replaces all user labels on the named resource.
func (h *Helper[T]) Set(ctx context.Context, name string, labels map[string]string) (T, error) {
	return h.svc.Patch(ctx, name, patch(labels))
}

// Add merges the given labels into the existing labels.
func (h *Helper[T]) Add(ctx context.Context, name string, add map[string]string) (T, error) {
	current, err := h.Get(ctx, name)
	if err != nil {
		var zero T
		return zero, err
	}
	return h.svc.Patch(ctx, name, patch(Merge(current, add, nil)))
}

// Remove removes the specified label keys.
func (h *Helper[T]) Remove(ctx context.Context, name string, keys []string) (T, error) {
	current, err := h.Get(ctx, name)
	if err != nil {
		var zero T
		return zero, err
	}
	return h.svc.Patch(ctx, name, patch(Merge(current, nil, keys)))
}

// Merge combines current labels with additions and removals into a new map.
func Merge(current, add map[string]string, remove []string) map[string]string {
	result := make(map[string]string, len(current)+len(add))
	for k, v := range current {
		result[k] = v
	}
	for k, v := range add {
		result[k] = v
	}
	for _, k := range remove {
		delete(result, k)
	}
	return result
}

func patch(labels map[string]string) map[string]interface{} {
	return map[string]interface{}{
		"metadata": map[string]interface{}{
			"userLabels": labels,
		},
	}
}
