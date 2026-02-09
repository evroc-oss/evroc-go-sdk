// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package rest provides HTTP client utilities for the evroc SDK.
//
// This package contains shared functionality used by all API service clients,
// including automatic retries, waiters, and HTTP request/response handling.
//
// # Automatic Retries
//
// The SDK automatically retries transient errors with exponential backoff:
//
//   - Network errors (timeouts, connection failures)
//   - Server errors (500, 502, 503, 504)
//   - Rate limiting (429)
//
// Default retry configuration:
//
//   - Maximum retries: 3
//   - Initial backoff: 1 second
//   - Maximum backoff: 30 seconds
//   - Backoff multiplier: 2.0
//   - Jitter: 25% (prevents thundering herd)
//
// Customize retry behavior:
//
//	config := rest.RetryConfig{
//	    MaxRetries:         5,
//	    InitialBackoff:     2 * time.Second,
//	    MaxBackoff:         60 * time.Second,
//	    BackoffMultiplier:  2.0,
//	    Jitter:             true,
//	}
//
// # Waiters
//
// Wait for resources to reach desired states with configurable polling:
//
//	config := rest.WaiterConfig{
//	    Timeout:        5 * time.Minute,
//	    InitialInterval: 2 * time.Second,
//	    MaxInterval:    30 * time.Second,
//	    Multiplier:     1.5,
//	    Exponential:    true,
//	}
//
//	err := rest.WaitFor(ctx, config, func() (bool, error) {
//	    resource, err := service.Get(ctx, name)
//	    if err != nil {
//	        return false, nil // Retry on errors
//	    }
//	    return resource.Status.Phase == "Ready", nil
//	})
//
// Monitor progress with callbacks:
//
//	err := rest.WaitFor(ctx, config, condition,
//	    rest.WithProgressCallback(func(attempt int, elapsed time.Duration) {
//	        log.Printf("Attempt %d, elapsed %s", attempt, elapsed)
//	    }),
//	)
//
// # Context Cancellation
//
// All operations respect context cancellation:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	// Waiter will stop when context times out
//	err := rest.WaitFor(ctx, config, condition)
//	if errors.Is(err, context.DeadlineExceeded) {
//	    log.Println("Operation timed out")
//	}
//
// # Error Handling
//
// The package provides utilities for checking retryable errors:
//
//	if rest.IsRetryable(err) {
//	    // Transient error, safe to retry
//	}
//
//	if rest.IsRateLimited(err) {
//	    // Rate limit exceeded
//	}
package rest
