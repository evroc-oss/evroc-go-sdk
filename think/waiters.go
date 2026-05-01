// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package think

import (
	"context"
	"errors"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/types/think"
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

// WaitForReady waits for an Instance to be in the Running state, at which point it should
// be ready to serve inference requests. Large models can take a significant time (10+ minutes) to be ready.
func (s *InstancesService) WaitForReady(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) (*think.Instance, error) {
	return s.waitForPhase(ctx, name, timeout, think.Running, opts...)
}

// WaitForStopped waits for an Instance to be in the Stopped state
func (s *InstancesService) WaitForStopped(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) (*think.Instance, error) {
	return s.waitForPhase(ctx, name, timeout, think.Stopped, opts...)
}

// WaitForDeleted polls until the instance returns 404 (deleted).
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *InstancesService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) error {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "instance"
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

func (s *InstancesService) waitForPhase(
	ctx context.Context,
	name string,
	timeout time.Duration,
	targetPhase think.InstanceStatusPhase,
	opts ...WaiterOption,
) (*think.Instance, error) {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "instance"
	config.Metrics = s.client.metrics

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	var result *think.Instance
	err := rest.WaitFor(ctx, config, func() (bool, error) {
		instance, err := s.Get(ctx, name)
		if err != nil {
			return false, nil
		}
		if instance == nil {
			return false, nil
		}
		var phase think.InstanceStatusPhase
		if instance.Status.Phase != nil {
			phase = *instance.Status.Phase
		}

		ready := phase == targetPhase

		if ready {
			result = instance
		}
		return ready, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}
