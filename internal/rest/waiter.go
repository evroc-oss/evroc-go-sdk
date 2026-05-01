// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"
)

var logger *slog.Logger

func init() {
	if os.Getenv("EVROC_SDK_DEBUG") != "" {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelError, // Only errors by default
		}))
	}
}

// Default waiter configuration constants.
const (
	// DefaultWaiterTimeout is the default maximum time to wait for a condition.
	DefaultWaiterTimeout = 5 * time.Minute

	// DefaultWaiterInitialInterval is the default initial polling interval.
	DefaultWaiterInitialInterval = 2 * time.Second

	// DefaultWaiterMaxInterval is the default maximum polling interval.
	DefaultWaiterMaxInterval = 30 * time.Second

	// DefaultWaiterMultiplier is the default multiplier for exponential backoff.
	// Set to 1.0 for constant interval polling.
	DefaultWaiterMultiplier = 1.5
)

// WaiterConfig defines configuration for polling operations.
type WaiterConfig struct {
	// Timeout is the maximum time to wait (default: 5 minutes)
	Timeout time.Duration

	// InitialInterval is the initial polling interval (default: 2s)
	InitialInterval time.Duration

	// MaxInterval is the maximum polling interval (default: 30s)
	MaxInterval time.Duration

	// Multiplier for exponential backoff (default: 1.5)
	// Set to 1.0 for constant interval
	Multiplier float64

	// ProgressCallback is called after each poll attempt (optional)
	// Receives attempt number and elapsed time
	ProgressCallback func(attempt int, elapsed time.Duration)

	// Metrics is an optional metrics recorder for waiter operations
	Metrics WaiterMetricsRecorder

	// ResourceType is the type of resource being waited on (for metrics labels)
	ResourceType string
}

// WaiterMetricsRecorder defines the interface for recording waiter metrics.
type WaiterMetricsRecorder interface {
	RecordWaiterOperation(resourceType string, duration float64, attempts int)
	RecordWaiterError(resourceType string, duration float64, attempts int)
}

// DefaultWaiterConfig returns sensible defaults for waiter operations.
func DefaultWaiterConfig() WaiterConfig {
	return WaiterConfig{
		Timeout:         DefaultWaiterTimeout,
		InitialInterval: DefaultWaiterInitialInterval,
		MaxInterval:     DefaultWaiterMaxInterval,
		Multiplier:      DefaultWaiterMultiplier,
	}
}

// WaitFor polls a condition function until it returns true or the context times out.
// The condition function is called repeatedly with exponential backoff.
// Returns nil if condition becomes true, error if timeout or context cancelled.
func WaitFor(ctx context.Context, config WaiterConfig, condition func() (done bool, err error)) error {
	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	startTime := time.Now()
	interval := config.InitialInterval
	attempt := 0

	for {
		attempt++
		elapsed := time.Since(startTime)

		// Call progress callback if provided
		if config.ProgressCallback != nil {
			config.ProgressCallback(attempt, elapsed)
		}

		// Debug logging
		logger.Debug("waiter polling",
			"attempt", attempt,
			"elapsed", elapsed)

		// Check condition
		done, err := condition()
		if err != nil {
			logger.Debug("waiter condition error",
				"error", err)

			// Record waiter error metrics
			if config.Metrics != nil && config.ResourceType != "" {
				duration := time.Since(startTime).Seconds()
				config.Metrics.RecordWaiterError(config.ResourceType, duration, attempt)
			}

			return fmt.Errorf("condition check failed: %w", err)
		}
		if done {
			logger.Debug("waiter condition satisfied",
				"attempts", attempt,
				"elapsed", elapsed)

			// Record successful waiter operation
			if config.Metrics != nil && config.ResourceType != "" {
				duration := time.Since(startTime).Seconds()
				config.Metrics.RecordWaiterOperation(config.ResourceType, duration, attempt)
			}

			return nil
		}

		logger.Debug("waiter condition not met",
			"interval", interval)

		// Wait for next poll with context awareness
		select {
		case <-time.After(interval):
			// Calculate next interval with exponential backoff
			interval = time.Duration(float64(interval) * config.Multiplier)
			if interval > config.MaxInterval {
				interval = config.MaxInterval
			}

			// Continue polling
		case <-ctx.Done():
			// Record waiter timeout/cancellation as error
			if config.Metrics != nil && config.ResourceType != "" {
				duration := time.Since(startTime).Seconds()
				config.Metrics.RecordWaiterError(config.ResourceType, duration, attempt)
			}

			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("timeout after %v (attempted %d times)", config.Timeout, attempt)
			}
			return fmt.Errorf("cancelled after %v: %w", elapsed, ctx.Err())
		}
	}
}

// WaitForWithConstantInterval is a convenience wrapper that uses constant polling intervals.
func WaitForWithConstantInterval(ctx context.Context, timeout, interval time.Duration, condition func() (bool, error)) error {
	config := WaiterConfig{
		Timeout:         timeout,
		InitialInterval: interval,
		MaxInterval:     interval,
		Multiplier:      1.0, // No backoff
	}
	return WaitFor(ctx, config, condition)
}
