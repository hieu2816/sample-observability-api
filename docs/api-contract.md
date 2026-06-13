# API Contract

All JSON responses use `Content-Type: application/json`.

## Health And Metadata

### `GET /health`

Returns `200` when the process is alive.

```json
{"status":"ok"}
```

### `GET /ready`

Returns `200` when runtime readiness is enabled and `503` when disabled.

```json
{"ready":true}
```

```json
{"ready":false,"reason":"readiness disabled by runtime config"}
```

### `GET /version`

Returns build metadata.

```json
{"version":"dev","commit":"unknown","build_time":"unknown"}
```

## Workload Endpoints

### `GET /api/orders`

Returns stable sample order data. This endpoint is affected by configured latency and error rate.

### `GET /api/payments`

Returns stable sample payment data. This endpoint is affected by configured latency and error rate.

## Runtime Config Endpoints

### `POST /config/error-rate`

```json
{"error_rate_percent":25}
```

Valid range is `0` to `100`. The alias `error_rate` is also accepted.

### `POST /config/latency`

```json
{"latency":"250ms"}
```

The alternative `{"latency_ms":250}` is also accepted. Maximum latency is `30s`.

### `POST /config/readiness`

```json
{"ready":false}
```

## Simulation Endpoints

### `GET /simulate/error`

Always returns `500`.

### `GET /simulate/latency?duration=2s`

Sleeps for the requested duration, then returns `200`. Maximum duration is `30s`.

### `GET /simulate/cpu-stress?duration=5s`

Runs a CPU-bound loop for the requested duration. Maximum duration is `10s`.

### `GET /simulate/memory-stress?mb=64&duration=10s`

Allocates and touches the requested number of megabytes, then keeps the allocation alive for the requested duration so resource monitors have time to observe memory pressure. Valid memory range is `1` to `256` MB. Maximum duration is `30s`; the default duration is `5s`.
