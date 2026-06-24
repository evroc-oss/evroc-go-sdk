// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package evroc

// NameFromRef extracts the resource name (last path segment) from a
// fully-qualified resource reference (FQID).
// For example, "/loadbalancer/projects/p/regions/r/backendPools/my-pool"
// returns "my-pool".
func NameFromRef(ref string) string {
	for i := len(ref) - 1; i >= 0; i-- {
		if ref[i] == '/' {
			return ref[i+1:]
		}
	}
	return ref
}
