// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package rest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryLogic(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		failureCount   int
		expectRetries  int
		expectSuccess  bool
		retryConfig    RetryConfig
	}{
		{
			name:          "success on first try",
			statusCode:    200,
			failureCount:  0,
			expectRetries: 0,
			expectSuccess: true,
			retryConfig:   DefaultRetryConfig(),
		},
		{
			name:          "success after 2 retries on 500",
			statusCode:    500,
			failureCount:  2,
			expectRetries: 2,
			expectSuccess: true,
			retryConfig:   DefaultRetryConfig(),
		},
		{
			name:          "fail after max retries on 503",
			statusCode:    503,
			failureCount:  5,
			expectRetries: 3, // Max retries
			expectSuccess: false,
			retryConfig:   DefaultRetryConfig(),
		},
		{
			name:          "no retry on 404",
			statusCode:    404,
			failureCount:  1,
			expectRetries: 0,
			expectSuccess: false,
			retryConfig:   DefaultRetryConfig(),
		},
		{
			name:          "retry on 429 rate limit",
			statusCode:    429,
			failureCount:  1,
			expectRetries: 1,
			expectSuccess: true,
			retryConfig:   DefaultRetryConfig(),
		},
		{
			name:          "custom retry config with 1 max retry",
			statusCode:    502,
			failureCount:  2,
			expectRetries: 1,
			expectSuccess: false,
			retryConfig: RetryConfig{
				MaxRetries:           1,
				InitialBackoff:       10 * time.Millisecond,
				MaxBackoff:           100 * time.Millisecond,
				BackoffMultiplier:    2.0,
				Jitter:               false,
				RetryableStatusCodes: []int{502},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attemptCount int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count := atomic.AddInt32(&attemptCount, 1)
				if int(count) <= tt.failureCount {
					w.WriteHeader(tt.statusCode)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := &Client{
				baseURL:     server.URL,
				httpClient:  server.Client(),
				retryConfig: tt.retryConfig,
			}

			req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
			ctx := context.Background()

			resp, err := client.Do(ctx, req)

			actualRetries := int(atomic.LoadInt32(&attemptCount)) - 1

			if tt.expectSuccess {
				if err != nil {
					t.Errorf("expected success, got error: %v", err)
				}
				if resp == nil || resp.StatusCode != 200 {
					t.Errorf("expected 200 response, got: %v", resp)
				}
			} else {
				if err == nil {
					t.Errorf("expected error, got success")
				}
			}

			if actualRetries != tt.expectRetries {
				t.Errorf("expected %d retries, got %d", tt.expectRetries, actualRetries)
			}
		})
	}
}

func TestRetryWithContextCancellation(t *testing.T) {
	var attemptCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attemptCount, 1)
		time.Sleep(50 * time.Millisecond) // Simulate slow response
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
		retryConfig: RetryConfig{
			MaxRetries:           5,
			InitialBackoff:       100 * time.Millisecond,
			MaxBackoff:           1 * time.Second,
			BackoffMultiplier:    2.0,
			Jitter:               false,
			RetryableStatusCodes: []int{503},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)

	start := time.Now()
	_, err := client.Do(ctx, req)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected error due to context cancellation")
	}

	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Errorf("expected context error, got: %v", err)
	}

	// Should not take much longer than the context timeout
	if elapsed > 500*time.Millisecond {
		t.Errorf("retry took too long after context cancellation: %v", elapsed)
	}

	attempts := int(atomic.LoadInt32(&attemptCount))
	if attempts > 3 {
		t.Errorf("too many attempts after context cancellation: %d", attempts)
	}
}

func TestRetryBackoffCalculation(t *testing.T) {
	tests := []struct {
		name     string
		config   RetryConfig
		attempt  int
		minTime  time.Duration
		maxTime  time.Duration
	}{
		{
			name: "exponential backoff without jitter",
			config: RetryConfig{
				InitialBackoff:    1 * time.Second,
				MaxBackoff:        30 * time.Second,
				BackoffMultiplier: 2.0,
				Jitter:            false,
			},
			attempt: 0,
			minTime: 1 * time.Second,
			maxTime: 1 * time.Second,
		},
		{
			name: "exponential backoff attempt 2",
			config: RetryConfig{
				InitialBackoff:    1 * time.Second,
				MaxBackoff:        30 * time.Second,
				BackoffMultiplier: 2.0,
				Jitter:            false,
			},
			attempt: 2,
			minTime: 4 * time.Second,
			maxTime: 4 * time.Second,
		},
		{
			name: "capped at max backoff",
			config: RetryConfig{
				InitialBackoff:    1 * time.Second,
				MaxBackoff:        5 * time.Second,
				BackoffMultiplier: 2.0,
				Jitter:            false,
			},
			attempt: 10,
			minTime: 5 * time.Second,
			maxTime: 5 * time.Second,
		},
		{
			name: "with jitter",
			config: RetryConfig{
				InitialBackoff:    1 * time.Second,
				MaxBackoff:        30 * time.Second,
				BackoffMultiplier: 2.0,
				Jitter:            true,
			},
			attempt: 1,
			minTime: 2 * time.Second,
			maxTime: 3 * time.Second, // 2s + 25% jitter = 2.5s max
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backoff := tt.config.calculateBackoff(tt.attempt)
			if backoff < tt.minTime || backoff > tt.maxTime {
				t.Errorf("backoff %v not in expected range [%v, %v]", backoff, tt.minTime, tt.maxTime)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "nil error",
			err:       nil,
			retryable: false,
		},
		{
			name:      "random error",
			err:       errors.New("some error"),
			retryable: false,
		},
		// Note: Testing actual network errors is difficult in unit tests
		// These would be integration test scenarios
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.retryable {
				t.Errorf("expected retryable=%v, got %v", tt.retryable, result)
			}
		})
	}
}
