# Metrics and Observability

The evroc Go SDK provides built-in Prometheus metrics to monitor API usage, performance, and errors.

## Available Metrics

The SDK automatically collects the following metrics when metrics are enabled:

| Metric Name | Type | Description | Labels |
|------------|------|-------------|--------|
| `evroc_sdk_api_calls_total` | Counter | Total number of API calls | `service`, `method`, `status_code` |
| `evroc_sdk_api_calls_duration_seconds` | Histogram | API call latency distribution | `service`, `method` |
| `evroc_sdk_api_calls_errors_total` | Counter | Total API errors | `service`, `method`, `error_type` |
| `evroc_sdk_retries_total` | Counter | Number of retry attempts | `service`, `method` |
| `evroc_sdk_waiter_operations_total` | Counter | Waiter operations (e.g., WaitForReady) | `service`, `resource_type`, `status` |
| `evroc_sdk_waiter_duration_seconds` | Histogram | Time spent waiting for resources | `service`, `resource_type` |
| `evroc_sdk_auth_token_refreshes_total` | Counter | OAuth token refresh operations | `status` |

## Enabling Metrics

### Basic Setup

```go
package main

import (
    "context"
    "log"
    "net/http"

    evroc "github.com/evroc-oss/evroc-go-sdk"
    "github.com/evroc-oss/evroc-go-sdk/metrics"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    ctx := context.Background()

    // 1. Create metrics manager
    metricsManager := metrics.NewManager()

    // 2. Start metrics HTTP server
    go func() {
        http.Handle("/metrics", promhttp.HandlerFor(
            metricsManager.Registry(),
            promhttp.HandlerOpts{},
        ))
        if err := http.ListenAndServe(":9090", nil); err != nil {
            log.Printf("Metrics server error: %v", err)
        }
    }()

    // 3. Create SDK client with metrics enabled
    client, err := evroc.NewFromEnv(ctx, evroc.WithMetrics(metricsManager))
    if err != nil {
        log.Fatal(err)
    }

    // Now all SDK operations will emit metrics
    // Access metrics at http://localhost:9090/metrics
}
```

### Using a Custom Registry

If you're already using Prometheus in your application:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    evrocmetrics "github.com/evroc-oss/evroc-go-sdk/metrics"
)

// Use your existing Prometheus registry
customRegistry := prometheus.NewRegistry()
metricsManager := evrocmetrics.NewManagerWithRegistry(customRegistry)

client, err := evroc.NewFromEnv(ctx, evroc.WithMetrics(metricsManager))
```

## Example Queries

Once metrics are exposed, you can query them with Prometheus or view them in Grafana:

### API Request Rate

```promql
# Requests per second by service
rate(evroc_sdk_api_calls_total[5m])

# Requests per second by service and method
rate(evroc_sdk_api_calls_total{service="compute"}[5m])
```

### Error Rate

```promql
# Error percentage
sum(rate(evroc_sdk_api_calls_errors_total[5m]))
/
sum(rate(evroc_sdk_api_calls_total[5m])) * 100
```

### Latency

```promql
# 95th percentile API latency
histogram_quantile(0.95,
  rate(evroc_sdk_api_calls_duration_seconds_bucket[5m])
)

# Average latency by service
rate(evroc_sdk_api_calls_duration_seconds_sum[5m])
/
rate(evroc_sdk_api_calls_duration_seconds_count[5m])
```

### Retry Rate

```promql
# Retry attempts per second
rate(evroc_sdk_retries_total[5m])

# Retry rate by service
sum by (service) (rate(evroc_sdk_retries_total[5m]))
```

### Waiter Performance

```promql
# Average time waiting for resources to be ready
rate(evroc_sdk_waiter_duration_seconds_sum[5m])
/
rate(evroc_sdk_waiter_duration_seconds_count[5m])

# Waiter success rate
sum(rate(evroc_sdk_waiter_operations_total{status="success"}[5m]))
/
sum(rate(evroc_sdk_waiter_operations_total[5m]))
```

## Complete Example

See the **[metrics example](../examples/metrics/)** for a complete working demonstration that:
- Sets up metrics collection
- Exposes metrics via HTTP endpoint
- Continuously generates metrics by creating/deleting resources
- Shows how to inspect metrics in real-time

Run the example:

```bash
cd examples/metrics
go run main.go

# In another terminal, view metrics:
curl http://localhost:9090/metrics | grep evroc_sdk
```

## Grafana Dashboard

Use the PromQL queries above to build a Grafana dashboard tailored to your environment.

### Key Panels

- **API Request Rate** - Requests per second by service
- **Error Rate** - Percentage of failed requests
- **Latency Percentiles** - P50, P95, P99 latencies
- **Retry Rate** - Retry attempts over time
- **Waiter Duration** - Time spent waiting for resources
- **Token Refreshes** - OAuth token refresh operations

## Best Practices

1. **Always enable metrics in production** - Essential for debugging and monitoring
2. **Set appropriate histogram buckets** - Default buckets work for most use cases
3. **Monitor error rates** - Set up alerts for elevated error rates
4. **Track retry rates** - High retry rates may indicate API issues
5. **Monitor waiter durations** - Helps identify slow resource provisioning

## Performance Impact

Metrics collection has minimal performance overhead:
- ~10 microseconds per metric recording
- ~100KB memory for default metric collectors
- No impact on API request latency

Metrics are safe to use in production environments.

## Troubleshooting

### Metrics not appearing

```go
// Ensure metrics manager is created
manager := metrics.NewManager()

// Ensure client uses metrics
client, err := evroc.NewFromEnv(ctx, evroc.WithMetrics(manager))

// Ensure HTTP server is running
http.Handle("/metrics", promhttp.HandlerFor(manager.Registry(), promhttp.HandlerOpts{}))
http.ListenAndServe(":9090", nil)
```

### Port already in use

Change the metrics port:

```go
http.ListenAndServe(":8080", nil)  // Use different port
```

### Metrics endpoint returns empty

Ensure you've made at least one API call before scraping metrics.
