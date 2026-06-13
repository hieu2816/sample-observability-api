# Error Rate Evidence - 2026-06-12

## Experiment

Source: `experiments/error-rate-test.md`

## Environment

- Mode: local
- Command: `go run ./cmd/api`
- Port: `8080`
- Runtime error rate configured: `50%`

## Commands Run

- `curl.exe -i -X POST http://localhost:8080/config/error-rate -H "Content-Type: application/json" --data-binary "@error-rate-50.json"`
- `1..10 | ForEach-Object { curl.exe -s -i http://localhost:8080/api/orders }`
- `1..10 | ForEach-Object { curl.exe -s -i http://localhost:8080/api/payments }`
- `curl.exe -s http://localhost:8080/metrics -o evidence\logs\error-rate-metrics-2026-06-12.txt`
- `curl.exe -i -X POST http://localhost:8080/config/error-rate -H "Content-Type: application/json" --data-binary "@error-rate-0.json"`

## Observed Results

| Route | Successful Responses | Injected 500 Responses |
| --- | ---: | ---: |
| `/api/orders` | 6 | 4 |
| `/api/payments` | 4 | 6 |

The response body for injected failures included:

- `{"error":"injected_error","message":"runtime error-rate configuration injected this response"}`

## Metrics Evidence

Raw metrics output:

- `evidence/logs/error-rate-metrics-2026-06-12.txt`

Important evidence lines:

- `sample_api_http_requests_total{method="GET",route="/api/orders",status="500"} 4`
- `sample_api_http_requests_total{method="GET",route="/api/payments",status="500"} 6`
- `sample_api_http_errors_total{method="GET",route="/api/orders",status="500"} 4`
- `sample_api_http_errors_total{method="GET",route="/api/payments",status="500"} 6`
- `sample_api_configured_error_rate_percent 50`

## Conclusion

The API can inject controlled 500 responses into workload endpoints using runtime error-rate configuration. Prometheus-compatible metrics captured the injected failures by route and status code, which is suitable for later SLO gate, canary rollback, and alerting experiments.

