// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package loadbalancer

import (
	"context"
	"errors"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	lbtypes "github.com/evroc-oss/evroc-go-sdk/types/loadbalancer"
)

// WaiterOption allows customizing waiter behavior.
type WaiterOption func(*rest.WaiterConfig)

// WithPollingInterval sets a constant polling interval (disables exponential backoff).
func WithPollingInterval(interval time.Duration) WaiterOption {
	return func(cfg *rest.WaiterConfig) {
		cfg.InitialInterval = interval
		cfg.MaxInterval = interval
		cfg.Multiplier = 1.0
	}
}

// WithExponentialBackoff configures exponential backoff for polling.
func WithExponentialBackoff(initial, max time.Duration, multiplier float64) WaiterOption {
	return func(cfg *rest.WaiterConfig) {
		cfg.InitialInterval = initial
		cfg.MaxInterval = max
		cfg.Multiplier = multiplier
	}
}

// WithProgressCallback sets a callback to track wait progress.
func WithProgressCallback(callback func(attempt int, elapsed time.Duration)) WaiterOption {
	return func(cfg *rest.WaiterConfig) {
		cfg.ProgressCallback = callback
	}
}

// WaitForReady polls the LoadBalancer until it reaches the Ready condition or the timeout is reached.
func (s *LoadBalancersService) WaitForReady(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) (*lbtypes.Loadbalancer, error) {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "loadbalancer"
	config.Metrics = s.client.metrics

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	var result *lbtypes.Loadbalancer
	err := rest.WaitFor(ctx, config, func() (bool, error) {
		lb, err := s.Get(ctx, name)
		if err != nil {
			return false, nil
		}
		if IsReady(lb) {
			result = lb
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// WaitForDeleted polls the LoadBalancer until it returns 404 (deleted) or the timeout is reached.
func (s *LoadBalancersService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) error {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "loadbalancer"
	config.Metrics = s.client.metrics

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	return rest.WaitFor(ctx, config, func() (bool, error) {
		_, err := s.Get(ctx, name)
		if errors.Is(err, rest.ErrNotFound) {
			return true, nil
		}
		return false, nil
	})
}

// WaitForDeleted polls the BackendPool until it returns 404 (deleted) or the timeout is reached.
func (s *BackendPoolsService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) error {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "backendpool"
	config.Metrics = s.client.metrics

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	return rest.WaitFor(ctx, config, func() (bool, error) {
		_, err := s.Get(ctx, name)
		if errors.Is(err, rest.ErrNotFound) {
			return true, nil
		}
		return false, nil
	})
}

// WaitForDeleted polls the BackendService until it returns 404 (deleted) or the timeout is reached.
func (s *BackendServicesService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) error {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "backendservice"
	config.Metrics = s.client.metrics

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	return rest.WaitFor(ctx, config, func() (bool, error) {
		_, err := s.Get(ctx, name)
		if errors.Is(err, rest.ErrNotFound) {
			return true, nil
		}
		return false, nil
	})
}

// WaitForDeleted polls the L4Route until it returns 404 (deleted) or the timeout is reached.
func (s *L4RoutesService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) error {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "l4route"
	config.Metrics = s.client.metrics

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	return rest.WaitFor(ctx, config, func() (bool, error) {
		_, err := s.Get(ctx, name)
		if errors.Is(err, rest.ErrNotFound) {
			return true, nil
		}
		return false, nil
	})
}
