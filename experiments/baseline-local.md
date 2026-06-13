# Baseline Local Experiment

## Goal

Measure the API baseline behavior before injecting latency or errors.

## Setup

```powershell
go run ./cmd/api
```

## Commands

```powershell
curl http://localhost:8080/health
curl http://localhost:8080/ready
curl http://localhost:8080/api/orders
curl http://localhost:8080/api/payments
curl http://localhost:8080/metrics
```

## Evidence To Store

Store command output, screenshots, or benchmark results under `evidence/`.

## Result

Not measured yet.
