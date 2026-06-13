# Error Rate Test

## Goal

Verify that configured error rate changes workload endpoint behavior and appears in metrics.

## Setup

```powershell
go run ./cmd/api
```

## Commands

```powershell
curl -X POST http://localhost:8080/config/error-rate -H "Content-Type: application/json" -d "{\"error_rate_percent\":50}"
curl http://localhost:8080/api/orders
curl http://localhost:8080/api/payments
curl http://localhost:8080/metrics
```

## Evidence To Store

Store request output and `/metrics` output under `evidence/benchmark-results/` or `evidence/logs/`.

## Result

Not measured yet.
