// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package metrics provides Prometheus metrics for the evroc Go SDK.
//
// This package enables optional instrumentation of SDK operations with Prometheus
// metrics. Metrics are collected for API calls, retries, and waiter operations.
//
// # Usage
//
// Create a metrics manager and pass it to the SDK client:
//
//	manager := metrics.NewManager()
//	client, err := evroc.NewFromEnv(ctx, evroc.WithMetrics(manager))
//
// Expose metrics endpoint:
//
//	http.Handle("/metrics", promhttp.HandlerFor(manager.Registry(), promhttp.HandlerOpts{}))
//	http.ListenAndServe(":9090", nil)
//
// # Available Metrics
//
// API Call Metrics:
//   - evroc_sdk_api_calls_total - Total API calls by method and status
//   - evroc_sdk_api_calls_duration_seconds - API call duration histogram
//   - evroc_sdk_api_calls_errors_total - API call errors by method and type
//
// Retry Metrics:
//   - evroc_sdk_retries_total - Total retry attempts by method
//   - evroc_sdk_retry_backoff_seconds - Retry backoff duration histogram
//
// Waiter Metrics:
//   - evroc_sdk_waiter_operations_total - Waiter operations by resource type and status
//   - evroc_sdk_waiter_duration_seconds - Waiter duration histogram
//   - evroc_sdk_waiter_attempts_total - Waiter polling attempts histogram
//
// Auth Metrics:
//   - evroc_sdk_auth_token_refreshes_total - Token refresh attempts by status
//   - evroc_sdk_auth_token_refresh_errors_total - Token refresh errors by type
//   - evroc_sdk_auth_token_refresh_duration_seconds - Token refresh duration histogram
//   - evroc_sdk_auth_initial_auth_total - Initial authentication attempts by status
//   - evroc_sdk_auth_initial_auth_errors_total - Initial authentication errors by type
//   - evroc_sdk_auth_initial_auth_duration_seconds - Initial authentication duration histogram
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const (
	namespace = "evroc_sdk"
)

// Manager handles Prometheus metrics collection for the evroc SDK.
type Manager struct {
	registry *prometheus.Registry

	// API call metrics
	APICallsTotal    *prometheus.CounterVec
	APICallsDuration *prometheus.HistogramVec
	APICallsErrors   *prometheus.CounterVec

	// Retry metrics
	RetriesTotal     *prometheus.CounterVec
	RetryBackoffTime *prometheus.HistogramVec

	// Waiter metrics
	WaiterOperationsTotal *prometheus.CounterVec
	WaiterDuration        *prometheus.HistogramVec
	WaiterAttempts        *prometheus.HistogramVec

	// Auth metrics
	AuthTokenRefreshesTotal *prometheus.CounterVec
	AuthTokenRefreshErrors  *prometheus.CounterVec
	AuthTokenRefreshTime    *prometheus.HistogramVec
	AuthInitialAuthTotal    *prometheus.CounterVec
	AuthInitialAuthErrors   *prometheus.CounterVec
	AuthInitialAuthTime     *prometheus.HistogramVec
}

// NewManager creates a new metrics manager with all metrics registered.
//
// The manager includes default Go runtime metrics (memory, goroutines) and
// process metrics (CPU, file descriptors).
func NewManager() *Manager {
	registry := prometheus.NewRegistry()

	// Register default Go metrics
	registry.MustRegister(collectors.NewGoCollector())

	// Register default process metrics
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	m := &Manager{
		registry: registry,

		// API calls counter
		APICallsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "api_calls_total",
				Help:      "Total number of API calls by method and status",
			},
			[]string{"method", "service", "status"},
		),

		// API calls duration histogram
		APICallsDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "api_calls_duration_seconds",
				Help:      "Duration of API calls in seconds",
				Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10},
			},
			[]string{"method", "service"},
		),

		// API calls errors
		APICallsErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "api_calls_errors_total",
				Help:      "Total number of API call errors by method and error type",
			},
			[]string{"method", "service", "error_type"},
		),

		// Retry attempts counter
		RetriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "retries_total",
				Help:      "Total number of retry attempts by method",
			},
			[]string{"method", "service"},
		),

		// Retry backoff duration
		RetryBackoffTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "retry_backoff_seconds",
				Help:      "Retry backoff duration in seconds",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
			},
			[]string{"method", "service"},
		),

		// Waiter operations counter
		WaiterOperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "waiter_operations_total",
				Help:      "Total number of waiter operations by resource type and status",
			},
			[]string{"resource_type", "status"},
		),

		// Waiter duration histogram
		WaiterDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "waiter_duration_seconds",
				Help:      "Duration of waiter operations in seconds",
				Buckets:   []float64{1, 5, 10, 30, 60, 120, 300, 600},
			},
			[]string{"resource_type"},
		),

		// Waiter attempts histogram
		WaiterAttempts: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "waiter_attempts_total",
				Help:      "Number of polling attempts during waiter operations",
				Buckets:   []float64{1, 2, 5, 10, 20, 50, 100},
			},
			[]string{"resource_type"},
		),

		// Auth token refresh counter
		AuthTokenRefreshesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_token_refreshes_total",
				Help:      "Total number of token refresh attempts",
			},
			[]string{"status"},
		),

		// Auth token refresh errors
		AuthTokenRefreshErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_token_refresh_errors_total",
				Help:      "Total number of token refresh errors",
			},
			[]string{"error_type"},
		),

		// Auth token refresh duration
		AuthTokenRefreshTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "auth_token_refresh_duration_seconds",
				Help:      "Duration of token refresh operations in seconds",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10},
			},
			[]string{},
		),

		// Auth initial authentication counter
		AuthInitialAuthTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_initial_auth_total",
				Help:      "Total number of initial authentication attempts",
			},
			[]string{"status"},
		),

		// Auth initial authentication errors
		AuthInitialAuthErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_initial_auth_errors_total",
				Help:      "Total number of initial authentication errors",
			},
			[]string{"error_type"},
		),

		// Auth initial authentication duration
		AuthInitialAuthTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "auth_initial_auth_duration_seconds",
				Help:      "Duration of initial authentication operations in seconds",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10},
			},
			[]string{},
		),
	}

	// Register all metrics
	registry.MustRegister(
		m.APICallsTotal,
		m.APICallsDuration,
		m.APICallsErrors,
		m.RetriesTotal,
		m.RetryBackoffTime,
		m.WaiterOperationsTotal,
		m.WaiterDuration,
		m.WaiterAttempts,
		m.AuthTokenRefreshesTotal,
		m.AuthTokenRefreshErrors,
		m.AuthTokenRefreshTime,
		m.AuthInitialAuthTotal,
		m.AuthInitialAuthErrors,
		m.AuthInitialAuthTime,
	)

	return m
}

// NewNoOpManager creates a no-op metrics manager for testing.
// All metrics are initialized but not registered, so writes are discarded.
func NewNoOpManager() *Manager {
	return &Manager{
		registry:                prometheus.NewRegistry(),
		APICallsTotal:           prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"method", "service", "status"}),
		APICallsDuration:        prometheus.NewHistogramVec(prometheus.HistogramOpts{}, []string{"method", "service"}),
		APICallsErrors:          prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"method", "service", "error_type"}),
		RetriesTotal:            prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"method", "service"}),
		RetryBackoffTime:        prometheus.NewHistogramVec(prometheus.HistogramOpts{}, []string{"method", "service"}),
		WaiterOperationsTotal:   prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"resource_type", "status"}),
		WaiterDuration:          prometheus.NewHistogramVec(prometheus.HistogramOpts{}, []string{"resource_type"}),
		WaiterAttempts:          prometheus.NewHistogramVec(prometheus.HistogramOpts{}, []string{"resource_type"}),
		AuthTokenRefreshesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"status"}),
		AuthTokenRefreshErrors:  prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"error_type"}),
		AuthTokenRefreshTime:    prometheus.NewHistogramVec(prometheus.HistogramOpts{}, []string{}),
		AuthInitialAuthTotal:    prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"status"}),
		AuthInitialAuthErrors:   prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"error_type"}),
		AuthInitialAuthTime:     prometheus.NewHistogramVec(prometheus.HistogramOpts{}, []string{}),
	}
}

// Registry returns the Prometheus registry.
// Use this to expose metrics via HTTP:
//
//	http.Handle("/metrics", promhttp.HandlerFor(manager.Registry(), promhttp.HandlerOpts{}))
func (m *Manager) Registry() *prometheus.Registry {
	return m.registry
}

// RecordAPICall records a successful API call with its duration.
func (m *Manager) RecordAPICall(method, service string, duration float64) {
	if m == nil {
		return
	}
	m.APICallsDuration.WithLabelValues(method, service).Observe(duration)
	m.APICallsTotal.WithLabelValues(method, service, "success").Inc()
}

// RecordAPICallError records a failed API call with its duration and error type.
func (m *Manager) RecordAPICallError(method, service string, duration float64, errorType string) {
	if m == nil {
		return
	}
	m.APICallsDuration.WithLabelValues(method, service).Observe(duration)
	m.APICallsTotal.WithLabelValues(method, service, "failure").Inc()
	m.APICallsErrors.WithLabelValues(method, service, errorType).Inc()
}

// RecordRetry records a retry attempt.
func (m *Manager) RecordRetry(method, service string, backoffDuration float64) {
	if m == nil {
		return
	}
	m.RetriesTotal.WithLabelValues(method, service).Inc()
	m.RetryBackoffTime.WithLabelValues(method, service).Observe(backoffDuration)
}

// RecordWaiterOperation records a successful waiter operation.
func (m *Manager) RecordWaiterOperation(resourceType string, duration float64, attempts int) {
	if m == nil {
		return
	}
	m.WaiterDuration.WithLabelValues(resourceType).Observe(duration)
	m.WaiterAttempts.WithLabelValues(resourceType).Observe(float64(attempts))
	m.WaiterOperationsTotal.WithLabelValues(resourceType, "success").Inc()
}

// RecordWaiterError records a failed waiter operation.
func (m *Manager) RecordWaiterError(resourceType string, duration float64, attempts int) {
	if m == nil {
		return
	}
	m.WaiterDuration.WithLabelValues(resourceType).Observe(duration)
	m.WaiterAttempts.WithLabelValues(resourceType).Observe(float64(attempts))
	m.WaiterOperationsTotal.WithLabelValues(resourceType, "failure").Inc()
}

// RecordTokenRefresh records a successful token refresh.
func (m *Manager) RecordTokenRefresh(duration float64) {
	if m == nil {
		return
	}
	m.AuthTokenRefreshTime.WithLabelValues().Observe(duration)
	m.AuthTokenRefreshesTotal.WithLabelValues("success").Inc()
}

// RecordTokenRefreshError records a failed token refresh.
func (m *Manager) RecordTokenRefreshError(duration float64, errorType string) {
	if m == nil {
		return
	}
	m.AuthTokenRefreshTime.WithLabelValues().Observe(duration)
	m.AuthTokenRefreshesTotal.WithLabelValues("failure").Inc()
	m.AuthTokenRefreshErrors.WithLabelValues(errorType).Inc()
}

// RecordInitialAuth records a successful initial authentication.
func (m *Manager) RecordInitialAuth(duration float64) {
	if m == nil {
		return
	}
	m.AuthInitialAuthTime.WithLabelValues().Observe(duration)
	m.AuthInitialAuthTotal.WithLabelValues("success").Inc()
}

// RecordInitialAuthError records a failed initial authentication.
func (m *Manager) RecordInitialAuthError(duration float64, errorType string) {
	if m == nil {
		return
	}
	m.AuthInitialAuthTime.WithLabelValues().Observe(duration)
	m.AuthInitialAuthTotal.WithLabelValues("failure").Inc()
	m.AuthInitialAuthErrors.WithLabelValues(errorType).Inc()
}
