// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"path"

	"github.com/evroc-oss/evroc-go-sdk/types/iam"
)

// Revoke revokes a Serviceaccountcredential by name. This cannot be undone.
func (s *ServiceAccountCredentialsService) Revoke(ctx context.Context, name, reason string) error {
	resp, err := s.client.rest.Post(ctx, path.Join(s.resourcePath(name), "revoke"), &iam.RevokeRequest{Reason: reason})
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}
