// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package kms provides access to the evroc KMS v1 API.
//
// # Resources
//
//   - Keys: regional encryption keys, with lifecycle actions (Activate, Suspend,
//     Rotate, Trim, Import, RotateImported, AbortDeletion, AbortTrimming)
//
// # Usage
//
//	keys := client.KMS().Keys()
//	key, _ := keys.Get(ctx, "my-key")
//	_ = keys.Rotate(ctx, "my-key")
package kms
