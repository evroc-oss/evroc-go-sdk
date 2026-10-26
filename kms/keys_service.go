// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package kms

import (
	"context"
	"path"

	"github.com/evroc-oss/evroc-go-sdk/types/kms"
)

// Request bodies for key actions; aliased so callers keep stable names across regenerations.
type (
	ImportRequest         = kms.PostKmsV1ProjectsProjectIDRegionsRegionNameKeysKeyIDImportJSONBody
	RotateImportedRequest = kms.PostKmsV1ProjectsProjectIDRegionsRegionNameKeysKeyIDRotateImportedJSONBody
	TrimRequest           = kms.PostKmsV1ProjectsProjectIDRegionsRegionNameKeysKeyIDTrimJSONBody
)

// HashSHA256 is the only hash accepted by Import and RotateImported.
const HashSHA256 = "SHA256"

func (s *KeysService) action(ctx context.Context, name, action string, body any) error {
	p := path.Join(s.client.path.ResourcePath(
		s.client.parent.DefaultProject(),
		s.client.parent.DefaultRegion(),
		resourceKeys,
		name), action)
	resp, err := s.client.rest.Post(ctx, p, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

// Activate re-enables a suspended key.
func (s *KeysService) Activate(ctx context.Context, name string) error {
	return s.action(ctx, name, "activate", nil)
}

// Suspend disables a key until it is activated again.
func (s *KeysService) Suspend(ctx context.Context, name string) error {
	return s.action(ctx, name, "suspend", nil)
}

// AbortDeletion cancels a pending key deletion.
func (s *KeysService) AbortDeletion(ctx context.Context, name string) error {
	return s.action(ctx, name, "abortDeletion", nil)
}

// Trim schedules removal of key versions below minimumVersionToKeep.
func (s *KeysService) Trim(ctx context.Context, name string, minimumVersionToKeep int) error {
	return s.action(ctx, name, "trim", &TrimRequest{MinimumVersionToKeep: minimumVersionToKeep})
}

// AbortTrimming cancels a pending trim.
func (s *KeysService) AbortTrimming(ctx context.Context, name string) error {
	return s.action(ctx, name, "abortTrimming", nil)
}

// Rotate creates a new key version.
func (s *KeysService) Rotate(ctx context.Context, name string) error {
	return s.action(ctx, name, "rotate", nil)
}

// Import uploads externally generated key material, wrapped with the key's import wrapping key.
func (s *KeysService) Import(ctx context.Context, name string, req *ImportRequest) error {
	return s.action(ctx, name, "import", req)
}

// RotateImported creates a new key version from externally generated key material.
func (s *KeysService) RotateImported(ctx context.Context, name string, req *RotateImportedRequest) error {
	return s.action(ctx, name, "rotateImported", req)
}
