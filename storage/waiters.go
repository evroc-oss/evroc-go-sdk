// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	storage "github.com/evroc-oss/evroc-go-sdk/types/storage"
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

// WaitForReady polls the bucket until it has a Ready condition with status True.
// Returns the ready bucket resource.
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *BucketsService) WaitForReady(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) (*storage.Bucket, error) {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "bucket"
	config.Metrics = s.client.metrics

	// Apply custom options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	var result *storage.Bucket
	err := rest.WaitFor(ctx, config, func() (bool, error) {
		bucket, err := s.Get(ctx, name)
		if err != nil {
			logger.Debug("bucket get error during wait",
				"bucket", name,
				"error", err)
			// Transient errors during polling are not fatal
			// The waiter will continue polling until timeout
			return false, nil
		}
		ready := IsBucketReady(bucket)

		// Get ready condition status for logging
		readyCondition := "unknown"
		if bucket.Status.Conditions != nil {
			for _, cond := range *bucket.Status.Conditions {
				if cond.Type == "Ready" {
					readyCondition = string(cond.Status)
					break
				}
			}
		}

		logger.Debug("bucket status check",
			"bucket", name,
			"ready", ready,
			"ready_condition", readyCondition)

		if ready {
			result = bucket
		}
		return ready, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// WaitForReady polls the bucket service account until it has a Ready condition with status True.
// Returns the ready bucket service account resource.
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *BucketServiceAccountsService) WaitForReady(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) (*storage.BucketServiceAccount, error) {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "bucket_service_account"
	config.Metrics = s.client.metrics

	// Apply custom options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	var result *storage.BucketServiceAccount
	err := rest.WaitFor(ctx, config, func() (bool, error) {
		bsa, err := s.Get(ctx, name)
		if err != nil {
			logger.Debug("bucket service account get error during wait",
				"bucketServiceAccount", name,
				"error", err)
			// Transient errors during polling are not fatal
			return false, nil
		}
		ready := IsBucketServiceAccountReady(bsa)

		// Get ready condition status for logging
		readyCondition := "unknown"
		if bsa.Status.Conditions != nil {
			for _, cond := range *bsa.Status.Conditions {
				if cond.Type == "Ready" {
					readyCondition = string(cond.Status)
					break
				}
			}
		}

		logger.Debug("bucket service account status check",
			"bucketServiceAccount", name,
			"ready", ready,
			"ready_condition", readyCondition)

		if ready {
			result = bsa
		}
		return ready, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// WaitForAvailable polls the file store until its status is Available.
// Returns the available file store resource.
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *FileStoresService) WaitForAvailable(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) (*storage.Filestore, error) {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "file_store"
	config.Metrics = s.client.metrics

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	var result *storage.Filestore
	err := rest.WaitFor(ctx, config, func() (bool, error) {
		fs, err := s.Get(ctx, name)
		if err != nil {
			logger.Debug("file store get error during wait",
				"fileStore", name,
				"error", err)
			return false, nil
		}

		status := "unknown"
		if fs.Status.Status != nil {
			status = string(*fs.Status.Status)
		}

		logger.Debug("file store status check",
			"fileStore", name,
			"status", status)

		available := IsFileStoreAvailable(fs)
		if available {
			result = fs
		}
		return available, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// WaitForDeleted waits for a file store to be fully deleted.
func (s *FileStoresService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) error {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "file_store"
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
