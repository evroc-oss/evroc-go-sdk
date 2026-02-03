// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package evroc provides helper functions for working with the evroc SDK.
package evroc

// Ptr returns a pointer to any value.
// This is useful when you need to pass a pointer to a literal value
// when using direct construction instead of builders.
//
// Example:
//
//	name := evroc.Ptr("my-vm")
//	enabled := evroc.Ptr(true)
//	size := evroc.Ptr(int32(100))
func Ptr[T any](v T) *T {
	return &v
}
