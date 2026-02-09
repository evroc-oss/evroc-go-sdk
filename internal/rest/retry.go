// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package rest

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"
)

// Default retry configuration constants.
const (
	// DefaultMaxRetries is the default maximum number of retry attempts.
	DefaultMaxRetries = 3

	// DefaultInitialBackoff is the default initial backoff duration.
	DefaultInitialBackoff = 1 * time.Second

	// DefaultMaxBackoff is the default maximum backoff duration.
	DefaultMaxBackoff = 30 * time.Second

	// DefaultBackoffMultiplier is the default multiplier for exponential backoff.
	DefaultBackoffMultiplier = 2.0

	// JitterPercentage is the percentage of jitter to add (0.25 = 25%).
	JitterPercentage = 0.25
)

// Retryable HTTP status codes.
var (
	// RetryableServerErrors are 5xx status codes that indicate transient server errors.
	RetryableServerErrors = []int{
		http.StatusInternalServerError,  // 500
		http.StatusBadGateway,            // 502
		http.StatusServiceUnavailable,    // 503
		http.StatusGatewayTimeout,        // 504
	}

	// RetryableRateLimitStatus is the status code for rate limiting.
	RetryableRateLimitStatus = http.StatusTooManyRequests // 429
)

// RetryConfig defines retry behavior for HTTP requests.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (default: 3)
	MaxRetries int

	// InitialBackoff is the initial backoff duration (default: 1s)
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration (default: 30s)
	MaxBackoff time.Duration

	// BackoffMultiplier is the multiplier for exponential backoff (default: 2.0)
	BackoffMultiplier float64

	// Jitter adds randomness to backoff to avoid thundering herd (default: true)
	//
	// Why jitter matters: Without jitter, if 100 clients all fail at the same time,
	// they'll all retry at exactly the same time (1s, 2s, 4s...), creating traffic spikes.
	// Jitter spreads retries across time, reducing load spikes on recovering servers.
	Jitter bool

	// RetryableStatusCodes are HTTP status codes that trigger retries (default: 429, 500, 502, 503, 504)
	// 429 (Too Many Requests) - rate limiting, will be respected when implemented
	// 5xx - server errors that are typically transient
	RetryableStatusCodes []int
}

// DefaultRetryConfig returns sensible defaults for retry behavior.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        DefaultMaxRetries,
		InitialBackoff:    DefaultInitialBackoff,
		MaxBackoff:        DefaultMaxBackoff,
		BackoffMultiplier: DefaultBackoffMultiplier,
		Jitter:            true,
		RetryableStatusCodes: append(
			[]int{RetryableRateLimitStatus},
			RetryableServerErrors...,
		),
	}
}

// isRetryableError determines if an error should trigger a retry.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Network errors are retryable
	var netErr net.Error
	if errors.As(err, &netErr) {
		// Timeout errors are retryable
		if netErr.Timeout() {
			return true
		}
	}

	// Connection errors are retryable
	if errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.EPIPE) {
		return true
	}

	return false
}

// isRetryableStatusCode checks if the HTTP status code is retryable.
func (rc *RetryConfig) isRetryableStatusCode(statusCode int) bool {
	for _, code := range rc.RetryableStatusCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}

// calculateBackoff calculates the backoff duration for the given attempt.
func (rc *RetryConfig) calculateBackoff(attempt int) time.Duration {
	// Calculate exponential backoff
	backoff := float64(rc.InitialBackoff) * math.Pow(rc.BackoffMultiplier, float64(attempt))

	// Cap at max backoff
	if backoff > float64(rc.MaxBackoff) {
		backoff = float64(rc.MaxBackoff)
	}

	// Add jitter to avoid thundering herd
	if rc.Jitter {
		// Random jitter between 0% and JitterPercentage of backoff
		jitter := rand.Float64() * JitterPercentage * backoff
		backoff += jitter
	}

	return time.Duration(backoff)
}

// shouldRetry determines if a request should be retried based on the error and response.
func (rc *RetryConfig) shouldRetry(err error, resp *http.Response, attempt int) bool {
	// Don't retry if max retries exceeded
	if attempt >= rc.MaxRetries {
		return false
	}

	// Check for retryable errors
	if isRetryableError(err) {
		return true
	}

	// Check for retryable status codes
	if resp != nil && rc.isRetryableStatusCode(resp.StatusCode) {
		return true
	}

	return false
}

// doWithRetry executes an HTTP request with retry logic.
func (c *Client) doWithRetry(ctx context.Context, req *http.Request, retryConfig RetryConfig) (*http.Response, error) {
	var lastErr error
	var resp *http.Response

	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		// Check if context is cancelled before attempting
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Clone the request for retry (in case body was consumed)
		reqClone := req.Clone(ctx)

		// Attempt the request
		resp, lastErr = c.doOnce(ctx, reqClone)

		// Success - return immediately
		if lastErr == nil && resp != nil && resp.StatusCode < 400 {
			return resp, nil
		}

		// Check if we should retry
		if !retryConfig.shouldRetry(lastErr, resp, attempt) {
			// Not retryable or max retries reached
			break
		}

		// Calculate backoff
		backoff := retryConfig.calculateBackoff(attempt)

		// Record retry metrics
		if c.metrics != nil {
			service := extractServiceName(req.URL.Path)
			c.metrics.RecordRetry(req.Method, service, backoff.Seconds())
		}

		// Wait with context awareness
		select {
		case <-time.After(backoff):
			// Continue to next retry
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// All retries exhausted
	if lastErr != nil {
		return nil, lastErr
	}

	return resp, nil
}

// extractServiceName extracts service name from URL path (e.g., /compute/v1/... -> compute)
func extractServiceName(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return "unknown"
}

// classifyAPIError classifies an API error for metrics labeling
func classifyAPIError(err error) string {
	if err == nil {
		return "none"
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.StatusCode == http.StatusBadRequest:
			return "bad_request"
		case apiErr.StatusCode == http.StatusUnauthorized:
			return "unauthorized"
		case apiErr.StatusCode == http.StatusForbidden:
			return "forbidden"
		case apiErr.StatusCode == http.StatusNotFound:
			return "not_found"
		case apiErr.StatusCode == http.StatusConflict:
			return "conflict"
		case apiErr.StatusCode == http.StatusTooManyRequests:
			return "rate_limit"
		case apiErr.StatusCode >= 500:
			return "server_error"
		default:
			return "client_error"
		}
	}

	return "unknown_error"
}

// doOnce performs a single HTTP request attempt without retries.
func (c *Client) doOnce(ctx context.Context, req *http.Request) (*http.Response, error) {
	if req.Header.Get("Content-Type") == "" && req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}

	// Extract service name from URL path (e.g., /compute/v1/... -> compute)
	service := extractServiceName(req.URL.Path)
	method := req.Method

	start := time.Now()
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	duration := time.Since(start).Seconds()

	if err != nil {
		// Record network/connection error
		if c.metrics != nil {
			c.metrics.RecordAPICallError(method, service, duration, "network_error")
		}
		return nil, err
	}

	if err := checkResponse(resp); err != nil {
		// Record API error
		if c.metrics != nil {
			errorType := classifyAPIError(err)
			c.metrics.RecordAPICallError(method, service, duration, errorType)
		}
		return resp, err
	}

	// Record successful API call
	if c.metrics != nil {
		c.metrics.RecordAPICall(method, service, duration)
	}

	return resp, nil
}
