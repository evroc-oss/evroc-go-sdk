// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	storagetypes "github.com/evroc-oss/evroc-go-sdk/types/storage"
)

// Re-export commonly used constants from types/storage for easier access.
// This allows users to write storage.Versioned instead of storagetypes.Versioned.

// Object Retention Mode constants
const (
	// Disabled means no version history is kept for objects, and objects cannot be locked.
	Disabled = string(storagetypes.Disabled)

	// Suspended enables versioning but objects can still be deleted.
	Suspended = string(storagetypes.Suspended)

	// Versioned enables full versioning without object locking.
	Versioned = string(storagetypes.Versioned)

	// Locking enables object locking features. Objects can be locked for a duration
	// from being modified and/or deleted.
	Locking = string(storagetypes.Locking)
)

// Object Locking Mode constants
const (
	// Immutable (COMPLIANCE mode) - Objects cannot be deleted or overwritten,
	// even by root users, until retention period expires.
	Immutable = string(storagetypes.Immutable)

	// Soft (GOVERNANCE mode) - Objects are protected but users with special
	// permissions can override the retention.
	Soft = string(storagetypes.Soft)
)

// Type aliases for shorter, more convenient type names
type (
	// RetentionMode is an alias for BucketSpecObjectRetentionMode
	RetentionMode = storagetypes.RetentionMode

	// LockingMode is an alias for BucketSpecDefaultObjectLockingMode
	LockingMode = storagetypes.LockingMode
)
