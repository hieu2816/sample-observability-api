# Runtime Configuration

Runtime config is stored in memory. It resets when the process restarts.

## Error Rate

```powershell
curl -X POST http://localhost:8080/config/error-rate `
  -H "Content-Type: application/json" `
  -d "{\"error_rate_percent\":50}"
```

This affects `/api/orders` and `/api/payments`.

## Artificial Latency

```powershell
curl -X POST http://localhost:8080/config/latency `
  -H "Content-Type: application/json" `
  -d "{\"latency\":\"1s\"}"
```

This affects `/api/orders` and `/api/payments`.

## Readiness

```powershell
curl -X POST http://localhost:8080/config/readiness `
  -H "Content-Type: application/json" `
  -d "{\"ready\":false}"
```

This makes `/ready` return `503`, which is useful for Kubernetes readiness experiments.

## Reset

```powershell
curl -X POST http://localhost:8080/config/error-rate -H "Content-Type: application/json" -d "{\"error_rate_percent\":0}"
curl -X POST http://localhost:8080/config/latency -H "Content-Type: application/json" -d "{\"latency\":\"0s\"}"
curl -X POST http://localhost:8080/config/readiness -H "Content-Type: application/json" -d "{\"ready\":true}"
```
