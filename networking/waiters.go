// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/types/networking"
)

var logger *slog.Logger

func init() {
	level := slog.LevelError
	if os.Getenv("EVROC_SDK_DEBUG") != "" {
		level = slog.LevelDebug
	}
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}

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

// WaitForReady polls the public IP until it has a Ready condition with status True.
// Returns the ready public IP resource.
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *PublicIPsService) WaitForReady(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) (*networking.PublicIP, error) {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "public_ip"
	config.Metrics = s.client.metrics

	// Apply custom options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	var result *networking.PublicIP
	err := rest.WaitFor(ctx, config, func() (bool, error) {
		publicIP, err := s.Get(ctx, name)
		if err != nil {
			logger.Debug("public IP get error during wait",
				"publicIP", name,
				"error", err)
			// Transient errors during polling are not fatal
			// The waiter will continue polling until timeout
			return false, nil
		}
		ready := IsPublicIPReady(publicIP)

		// Get ready condition status for logging
		readyCondition := "unknown"
		if publicIP.Status.Conditions != nil {
			for _, cond := range *publicIP.Status.Conditions {
				if cond.Type == "Ready" {
					readyCondition = string(cond.Status)
					break
				}
			}
		}

		logger.Debug("public IP status check",
			"publicIP", name,
			"ready", ready,
			"ready_condition", readyCondition)

		if ready {
			result = publicIP
		}
		return ready, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// WaitForReady polls the security group until it has a Ready condition with status True.
// Returns the ready security group resource.
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *SecurityGroupsService) WaitForReady(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) (*networking.SecurityGroup, error) {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "security_group"
	config.Metrics = s.client.metrics

	// Apply custom options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	var result *networking.SecurityGroup
	err := rest.WaitFor(ctx, config, func() (bool, error) {
		sg, err := s.Get(ctx, name)
		if err != nil {
			logger.Debug("security group get error during wait",
				"securityGroup", name,
				"error", err)
			// Transient errors during polling are not fatal
			return false, nil
		}
		ready := IsSecurityGroupReady(sg)

		// Get ready condition status for logging
		readyCondition := "unknown"
		if sg.Status.Conditions != nil {
			for _, cond := range *sg.Status.Conditions {
				if cond.Type == "Ready" {
					readyCondition = string(cond.Status)
					break
				}
			}
		}

		logger.Debug("security group status check",
			"securityGroup", name,
			"ready", ready,
			"ready_condition", readyCondition)

		if ready {
			result = sg
		}
		return ready, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// WaitForDeleted polls until the public IP returns 404 (deleted).
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *PublicIPsService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) error {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "public_ip"
	config.Metrics = s.client.metrics

	// Apply custom options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	return rest.WaitFor(ctx, config, func() (bool, error) {
		_, err := s.Get(ctx, name)
		if errors.Is(err, rest.ErrNotFound) {
			// Resource deleted successfully
			return true, nil
		}
		// Still exists, keep polling
		return false, nil
	})
}

// WaitForDeleted polls until the security group returns 404 (deleted).
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *SecurityGroupsService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) error {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "security_group"
	config.Metrics = s.client.metrics

	// Apply custom options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	return rest.WaitFor(ctx, config, func() (bool, error) {
		_, err := s.Get(ctx, name)
		if errors.Is(err, rest.ErrNotFound) {
			// Resource deleted successfully
			return true, nil
		}
		// Still exists, keep polling
		return false, nil
	})
}
