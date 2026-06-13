# Local Baseline Evidence - 2026-06-12

## Experiment

Source: `experiments/baseline-local.md`

## Environment

- Mode: local
- Command: `go run ./cmd/api`
- Port: `8080`
- Version endpoint result: `dev / unknown / unknown`

## Commands Run

- `curl http://localhost:8080/health`
- `curl http://localhost:8080/ready`
- `curl http://localhost:8080/version`
- `curl http://localhost:8080/api/orders`
- `curl http://localhost:8080/api/payments`
- `curl http://localhost:8080/metrics`

## Observed Results

| Endpoint | Status | Result |
| --- | --- | --- |
| `/health` | 200 | Returned `{"status":"ok"}` |
| `/ready` | 200 | Returned `{"ready":true}` |
| `/version` | 200 | Returned `{"version":"dev","commit":"unknown","build_time":"unknown"}` |
| `/api/orders` | 200 | Returned sample order payload |
| `/api/payments` | 200 | Returned sample payment payload |
| `/metrics` | 200 | Returned Prometheus-compatible metrics |

## Metrics Evidence

Raw metrics output:

- `evidence/logs/local-baseline-metrics-2026-06-12.txt`

Important baseline signals:

- `sample_api_ready 1`
- `sample_api_configured_error_rate_percent 0`
- `sample_api_configured_latency_seconds 0`
- `sample_api_build_info{version="dev",commit="unknown",build_time="unknown"} 1`

## Conclusion

The API starts successfully in local mode, exposes baseline health/readiness/version/workload endpoints, returns JSON responses, includes request IDs, and exposes Prometheus-compatible metrics.

