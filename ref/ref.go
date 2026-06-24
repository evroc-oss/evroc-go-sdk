// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package ref provides utilities for working with fully-qualified resource references (FQIDs).
package ref

// NameFromRef extracts the resource name (last path segment) from a
// fully-qualified resource reference (FQID).
// For example, "/loadbalancer/projects/p/regions/r/backendPools/my-pool"
// returns "my-pool".
func NameFromRef(r string) string {
	for i := len(r) - 1; i >= 0; i-- {
		if r[i] == '/' {
			return r[i+1:]
		}
	}
	return r
}
