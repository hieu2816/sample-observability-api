# evidence

Use this folder for measured API evidence.

Examples:

- local benchmark results
- endpoint response screenshots
- metrics output samples
- latency and error-rate test results
- logs from simulated incidents

Only record evidence that was actually measured or captured.

Suggested structure:

```text
evidence/
  benchmark-results/
  logs/
  reports/
  screenshots/
```

Each evidence file should mention:

- date/time of the run
- command or tool used
- API version or commit if available
- expected behavior
- actual result
- where raw output is stored
