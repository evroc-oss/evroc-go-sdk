// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// CatalogCache holds API catalog snapshots loaded on first validation.
// Successful loads are kept indefinitely; failed loads may be retried.
// The zero value is ready to use. Share only between clients with the same
// endpoint and credentials. There is no automatic expiry or refresh.
type CatalogCache struct {
	profiles catalogSnapshot
	images   catalogSnapshot
}

type catalogSnapshot struct {
	mu    sync.Mutex
	names map[string]bool
}

// WithCatalogCache shares catalog snapshots between clients. Configure it before
// using the client concurrently. By default, each client owns its own cache.
func (c *Client) WithCatalogCache(cache *CatalogCache) *Client {
	if cache != nil {
		c.catalogs = cache
	}
	return c
}

func (c *catalogSnapshot) validate(name string, load func() ([]string, error)) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name must not be empty")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.names == nil {
		names, err := load()
		if err != nil {
			return err
		}
		c.names = make(map[string]bool, len(names))
		for _, name := range names {
			c.names[name] = true
		}
	}
	if !c.names[name] {
		return fmt.Errorf("%q is not in the API catalog", name)
	}
	return nil
}

// Validate accepts a compute profile name or full reference and checks the catalog, fetched once on
// first use. API errors are wrapped and do not populate the cache.
func (s *ComputeProfilesService) Validate(ctx context.Context, nameOrRef string) error {
	name := strings.TrimPrefix(nameOrRef, "/compute/global/computeProfiles/")
	err := s.client.catalogs.profiles.validate(name, func() ([]string, error) {
		catalog, err := s.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("list catalog: %w", err)
		}
		if catalog == nil {
			return nil, fmt.Errorf("API returned a null catalog")
		}
		names := make([]string, 0, len(catalog.Items))
		for _, item := range catalog.Items {
			names = append(names, item.Metadata.Id)
		}
		return names, nil
	})
	if err != nil {
		return fmt.Errorf("compute profile: %w", err)
	}
	return nil
}

// Validate accepts a stock disk image name or full reference and checks the catalog, fetched once on
// first use. API errors are wrapped and do not populate the cache.
func (s *DiskImagesService) Validate(ctx context.Context, nameOrRef string) error {
	name := strings.TrimPrefix(nameOrRef, "/compute/global/diskImages/evroc/")
	err := s.client.catalogs.images.validate(name, func() ([]string, error) {
		catalog, err := s.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("list catalog: %w", err)
		}
		if catalog == nil {
			return nil, fmt.Errorf("API returned a null catalog")
		}
		names := make([]string, 0, len(catalog.Items))
		for _, item := range catalog.Items {
			names = append(names, item.Metadata.Id)
		}
		return names, nil
	})
	if err != nil {
		return fmt.Errorf("stock disk image: %w", err)
	}
	return nil
}
