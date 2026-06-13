# Logging Format

The API writes structured JSON logs to stdout.

## Request Log Fields

Each HTTP request produces an `http_request` log entry with:

- `time`
- `level`
- `msg`
- `request_id`
- `method`
- `path`
- `route`
- `status`
- `duration_ms`

The response also includes the `X-Request-ID` header. If the request already provides `X-Request-ID`, the API reuses it.

## Runtime Config Logs

Config endpoints produce `runtime_config_changed` logs with:

- `setting`
- `value`

These logs help connect an experiment action to later latency, error, or readiness behavior.

## Simulation Logs

Simulation endpoints produce `simulation_triggered` logs with the simulation type and relevant parameters.
