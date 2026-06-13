# Metrics Reference

`GET /metrics` exposes Prometheus-compatible text metrics.

## Request Metrics

- `sample_api_http_requests_total{method,route,status}` counts HTTP responses.
- `sample_api_http_errors_total{method,route,status}` counts HTTP 5xx responses.
- `sample_api_http_request_duration_seconds_bucket{method,route,le}` records request latency histogram buckets.
- `sample_api_http_request_duration_seconds_sum{method,route}` records total observed latency.
- `sample_api_http_request_duration_seconds_count{method,route}` records observed request count for the histogram.

These metrics support Kubernetes release gates such as error rate by route, p95 latency by route, and availability of workload endpoints during rollout.

## Runtime Config Metrics

- `sample_api_ready` is `1` when ready and `0` when not ready.
- `sample_api_configured_error_rate_percent` shows the active error injection rate.
- `sample_api_configured_latency_seconds` shows the active artificial latency.

These metrics explain whether a failure is caused by an intentional experiment.

## Build Metric

- `sample_api_build_info{version,commit,build_time}` exposes build metadata as labels with value `1`.

This metric is useful during rollout and rollback experiments because dashboards can show which version is serving traffic.
