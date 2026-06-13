# Readiness Failure Evidence - 2026-06-12

## Experiment

Source: `experiments/readiness-failure-test.md`

## Environment

- Mode: local
- Command: `go run ./cmd/api`
- Port: `8080`

## Commands Run

- `curl.exe -i http://localhost:8080/ready`
- `curl.exe -i -X POST http://localhost:8080/config/readiness -H "Content-Type: application/json" --data-binary "@readiness-false.json"`
- `curl.exe -i http://localhost:8080/ready`
- `curl.exe -i -X POST http://localhost:8080/config/readiness -H "Content-Type: application/json" --data-binary "@readiness-true.json"`
- `curl.exe -i http://localhost:8080/ready`
- `curl.exe -s http://localhost:8080/metrics -o evidence\logs\readiness-failure-metrics-2026-06-12.txt`

## Observed Results

| Step | Expected | Observed |
| --- | --- | --- |
| Initial readiness check | `/ready` returns 200 | Passed |
| Set readiness false | Config endpoint returns 200 | Passed |
| Readiness failure check | `/ready` returns 503 | Passed |
| Reset readiness true | Config endpoint returns 200 | Passed |
| Final readiness check | `/ready` returns 200 | Passed |

## Metrics Evidence

Raw metrics output:

- `evidence/logs/readiness-failure-metrics-2026-06-12.txt`

Important evidence lines:

- `sample_api_http_errors_total{method="GET",route="/ready",status="503"} 1`
- `sample_api_ready 1`

`sample_api_ready 1` was captured after readiness was reset to healthy. The `status="503"` request metric confirms that readiness failure occurred during the experiment.

## Conclusion

The API can switch readiness from healthy to unavailable and back to healthy without restarting. This behavior is suitable for Kubernetes readiness probe and rollout failure experiments.

