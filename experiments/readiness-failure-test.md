# Readiness Failure Test

## Goal

Verify that readiness can be toggled without restarting the API.

## Setup

```powershell
go run ./cmd/api
```

## Commands

```powershell
curl http://localhost:8080/ready
curl -X POST http://localhost:8080/config/readiness -H "Content-Type: application/json" -d "{\"ready\":false}"
curl http://localhost:8080/ready
curl -X POST http://localhost:8080/config/readiness -H "Content-Type: application/json" -d "{\"ready\":true}"
curl http://localhost:8080/ready
```

## Evidence To Store

Store request output and timestamps under `evidence/reports/`.

## Result

Not measured yet.
