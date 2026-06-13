# Sample Observability API

## Context

This repository contains a Go-based technical API used as the workload for DevOps, SRE, Kubernetes, and cloud operations experiments.

The application is intentionally not business-heavy. Its purpose is to generate measurable operational signals such as latency, error rate, readiness state, resource pressure, logs, version metadata, and Prometheus-compatible metrics.

## Goals

- Provide a stable workload for Kubernetes release experiments.
- Provide a reusable workload for AWS cloud operations automation.
- Expose health, readiness, version, and metrics endpoints.
- Support controlled failure simulation.
- Produce evidence for latency, error rate, availability, and rollback behavior.
- Keep the code readable enough to learn Go HTTP fundamentals.

## Install Go

Install Go 1.26 or newer from:

```text
https://go.dev/dl/
```

Then verify:

```powershell
go version
```

If PowerShell outside VS Code sees Go but the VS Code terminal does not, restart VS Code or run:

```powershell
$env:Path += ";C:\Program Files\Go\bin"
go version
```

## Run Locally

```powershell
go test ./...
go run ./cmd/api
```

The API listens on port `8080` by default. Override it with:

```powershell
$env:PORT="9090"
go run ./cmd/api
```

Smoke test from another terminal:

```powershell
curl http://localhost:8080/health
curl http://localhost:8080/ready
curl http://localhost:8080/version
curl http://localhost:8080/api/orders
curl http://localhost:8080/metrics
```

## Run With Docker

```powershell
docker build -t sample-observability-api:local .
docker run --rm -p 8080:8080 sample-observability-api:local
```

Build with version metadata:

```powershell
docker build `
  --build-arg VERSION=0.1.0 `
  --build-arg COMMIT=local `
  --build-arg BUILD_TIME=2026-06-11T00:00:00Z `
  -t sample-observability-api:0.1.0 .
```

## Runtime Configuration

Inject a 25 percent error rate into workload endpoints:

```powershell
curl -X POST http://localhost:8080/config/error-rate `
  -H "Content-Type: application/json" `
  -d "{\"error_rate_percent\":25}"
```

Add artificial latency to workload endpoints:

```powershell
curl -X POST http://localhost:8080/config/latency `
  -H "Content-Type: application/json" `
  -d "{\"latency\":\"250ms\"}"
```

Toggle readiness:

```powershell
curl -X POST http://localhost:8080/config/readiness `
  -H "Content-Type: application/json" `
  -d "{\"ready\":false}"
```

Reset to healthy behavior:

```powershell
curl -X POST http://localhost:8080/config/error-rate -H "Content-Type: application/json" -d "{\"error_rate_percent\":0}"
curl -X POST http://localhost:8080/config/latency -H "Content-Type: application/json" -d "{\"latency\":\"0s\"}"
curl -X POST http://localhost:8080/config/readiness -H "Content-Type: application/json" -d "{\"ready\":true}"
```

## Endpoint Summary

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/health` | Liveness-style health check. |
| GET | `/ready` | Readiness check controlled by runtime config. |
| GET | `/version` | Build/version metadata. |
| GET | `/metrics` | Prometheus-compatible metrics. |
| GET | `/api/orders` | Sample workload endpoint. |
| GET | `/api/payments` | Sample workload endpoint. |
| POST | `/config/error-rate` | Configure workload error injection. |
| POST | `/config/latency` | Configure workload latency injection. |
| POST | `/config/readiness` | Configure readiness state. |
| GET | `/simulate/error` | Intentional 500 for incident tests. |
| GET | `/simulate/latency` | Intentional latency with `duration` query parameter. |
| GET | `/simulate/cpu-stress` | CPU pressure with `duration` query parameter. |
| GET | `/simulate/memory-stress` | Memory pressure with `mb` and `duration` query parameters. |

## Documentation

- [API contract](docs/api-contract.md)
- [Metrics reference](docs/metrics.md)
- [Logging format](docs/logging.md)
- [Runtime configuration](docs/runtime-config.md)

## Evidence Rules

This repo should only store measured results. Use `experiments/` to describe how a test is run, then store raw output, screenshots, logs, or benchmark files under `evidence/`.
