package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	mathrand "math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sample-observability-api/internal/buildinfo"
	"sample-observability-api/internal/config"
	"sample-observability-api/internal/observability"
	"sample-observability-api/internal/simulate"
)

type ServerOptions struct {
	State   *config.State
	Metrics *observability.Recorder
	Logger  *slog.Logger
	Build   buildinfo.Info
}

type Server struct {
	state   *config.State
	metrics *observability.Recorder
	logger  *slog.Logger
	build   buildinfo.Info
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func NewServer(options ServerOptions) *Server {
	return &Server{
		state:   options.State,
		metrics: options.Metrics,
		logger:  options.Logger,
		build:   options.Build,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	s.route(mux, http.MethodGet, "/health", s.health, false)
	s.route(mux, http.MethodGet, "/ready", s.ready, false)
	s.route(mux, http.MethodGet, "/version", s.version, false)
	s.route(mux, http.MethodGet, "/metrics", s.metricsEndpoint, false)
	s.route(mux, http.MethodGet, "/api/orders", s.orders, true)
	s.route(mux, http.MethodGet, "/api/payments", s.payments, true)
	s.route(mux, http.MethodPost, "/config/error-rate", s.setErrorRate, false)
	s.route(mux, http.MethodPost, "/config/latency", s.setLatency, false)
	s.route(mux, http.MethodPost, "/config/readiness", s.setReadiness, false)
	s.route(mux, http.MethodGet, "/simulate/error", s.simulateError, false)
	s.route(mux, http.MethodGet, "/simulate/latency", s.simulateLatency, false)
	s.route(mux, http.MethodGet, "/simulate/cpu-stress", s.simulateCPUStress, false)
	s.route(mux, http.MethodGet, "/simulate/memory-stress", s.simulateMemoryStress, false)
	return mux
}

func (s *Server) route(mux *http.ServeMux, method string, path string, handler http.HandlerFunc, affectedByRuntimeConfig bool) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed for this endpoint")
			return
		}
		handler(w, r)
	})
	mux.Handle(path, s.middleware(method, path, affectedByRuntimeConfig, next))
}

func (s *Server) middleware(method string, route string, affectedByRuntimeConfig bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := requestIDFrom(r)
		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
		r = r.WithContext(ctx)
		w.Header().Set("X-Request-ID", requestID)

		recorder := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		snapshot := s.state.Snapshot()

		if affectedByRuntimeConfig && snapshot.Latency > 0 {
			time.Sleep(snapshot.Latency)
		}

		if affectedByRuntimeConfig && shouldInjectError(snapshot.ErrorRatePercent) {
			writeError(recorder, http.StatusInternalServerError, "injected_error", "runtime error-rate configuration injected this response")
		} else {
			next.ServeHTTP(recorder, r)
		}

		duration := time.Since(start)
		s.metrics.Observe(method, route, recorder.status, duration)
		s.logger.Info("http_request",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"route", route,
			"status", recorder.status,
			"duration_ms", duration.Milliseconds(),
		)
	})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, _ *http.Request) {
	snapshot := s.state.Snapshot()
	if !snapshot.Ready {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"ready":  false,
			"reason": "readiness disabled by runtime config",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ready": true})
}

func (s *Server) version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.build)
}

func (s *Server) metricsEndpoint(w http.ResponseWriter, r *http.Request) {
	s.metrics.Handler(s.state, s.build).ServeHTTP(w, r)
}

func (s *Server) orders(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"items": []map[string]any{
			{"id": "ord-1001", "status": "created", "total": 49.90},
			{"id": "ord-1002", "status": "paid", "total": 125.00},
		},
	})
}

func (s *Server) payments(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"items": []map[string]any{
			{"id": "pay-9001", "order_id": "ord-1001", "status": "authorized", "amount": 49.90},
			{"id": "pay-9002", "order_id": "ord-1002", "status": "captured", "amount": 125.00},
		},
	})
}

func (s *Server) setErrorRate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ErrorRatePercent *float64 `json:"error_rate_percent"`
		ErrorRate        *float64 `json:"error_rate"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	value := math.NaN()
	if body.ErrorRatePercent != nil {
		value = *body.ErrorRatePercent
	}
	if body.ErrorRate != nil {
		value = *body.ErrorRate
	}
	if math.IsNaN(value) {
		writeError(w, http.StatusBadRequest, "missing_error_rate", "provide error_rate_percent or error_rate")
		return
	}
	if err := s.state.SetErrorRatePercent(value); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_error_rate", err.Error())
		return
	}

	s.logger.Info("runtime_config_changed", "setting", "error_rate_percent", "value", value)
	writeJSON(w, http.StatusOK, s.state.Snapshot())
}

func (s *Server) setLatency(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Latency   string `json:"latency"`
		LatencyMS *int64 `json:"latency_ms"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	var value time.Duration
	switch {
	case body.Latency != "":
		parsed, err := time.ParseDuration(body.Latency)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_latency", "latency must use Go duration syntax such as 250ms or 2s")
			return
		}
		value = parsed
	case body.LatencyMS != nil:
		value = time.Duration(*body.LatencyMS) * time.Millisecond
	default:
		writeError(w, http.StatusBadRequest, "missing_latency", "provide latency or latency_ms")
		return
	}

	if err := s.state.SetLatency(value); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_latency", err.Error())
		return
	}

	s.logger.Info("runtime_config_changed", "setting", "latency", "value", value.String())
	writeJSON(w, http.StatusOK, s.state.Snapshot())
}

func (s *Server) setReadiness(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Ready *bool `json:"ready"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if body.Ready == nil {
		writeError(w, http.StatusBadRequest, "missing_ready", "provide ready as true or false")
		return
	}

	s.state.SetReady(*body.Ready)
	s.logger.Info("runtime_config_changed", "setting", "ready", "value", *body.Ready)
	writeJSON(w, http.StatusOK, s.state.Snapshot())
}

func (s *Server) simulateError(w http.ResponseWriter, _ *http.Request) {
	s.logger.Warn("simulation_triggered", "type", "error")
	writeError(w, http.StatusInternalServerError, "simulated_error", "intentional error for incident simulation")
}

func (s *Server) simulateLatency(w http.ResponseWriter, r *http.Request) {
	duration := 500 * time.Millisecond
	if value := r.URL.Query().Get("duration"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_duration", "duration must use Go duration syntax such as 500ms or 2s")
			return
		}
		duration = parsed
	}
	if duration < 0 || duration > 30*time.Second {
		writeError(w, http.StatusBadRequest, "invalid_duration", "duration must be between 0 and 30s")
		return
	}

	s.logger.Warn("simulation_triggered", "type", "latency", "duration", duration.String())
	time.Sleep(duration)
	writeJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"duration": duration.String(),
	})
}

func (s *Server) simulateCPUStress(w http.ResponseWriter, r *http.Request) {
	duration := 2 * time.Second
	if value := r.URL.Query().Get("duration"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_duration", "duration must use Go duration syntax such as 500ms or 2s")
			return
		}
		duration = parsed
	}
	if duration < 0 || duration > 10*time.Second {
		writeError(w, http.StatusBadRequest, "invalid_duration", "duration must be between 0 and 10s")
		return
	}

	s.logger.Warn("simulation_triggered", "type", "cpu-stress", "duration", duration.String())
	iterations := simulate.CPUStress(duration)
	writeJSON(w, http.StatusOK, map[string]any{
		"status":     "ok",
		"duration":   duration.String(),
		"iterations": iterations,
	})
}

func (s *Server) simulateMemoryStress(w http.ResponseWriter, r *http.Request) {
	megabytes := 32
	duration := 5 * time.Second
	if value := r.URL.Query().Get("mb"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_mb", "mb must be an integer")
			return
		}
		megabytes = parsed
	}
	if megabytes < 1 || megabytes > 256 {
		writeError(w, http.StatusBadRequest, "invalid_mb", "mb must be between 1 and 256")
		return
	}
	if value := r.URL.Query().Get("duration"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_duration", "duration must use Go duration syntax such as 500ms or 2s")
			return
		}
		duration = parsed
	}
	if duration < 0 || duration > 30*time.Second {
		writeError(w, http.StatusBadRequest, "invalid_duration", "duration must be between 0 and 30s")
		return
	}

	s.logger.Warn("simulation_triggered", "type", "memory-stress", "mb", megabytes, "duration", duration.String())
	bytesAllocated := simulate.MemoryStress(megabytes, duration)
	writeJSON(w, http.StatusOK, map[string]any{
		"status":          "ok",
		"megabytes":       megabytes,
		"duration":        duration.String(),
		"bytes_allocated": bytesAllocated,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.Decode(&struct{}{}) == nil {
		return errors.New("request body must contain only one JSON object")
	}
	return nil
}

func shouldInjectError(errorRatePercent float64) bool {
	if errorRatePercent <= 0 {
		return false
	}
	if errorRatePercent >= 100 {
		return true
	}
	return mathrand.Float64()*100 < errorRatePercent
}

type requestIDKey struct{}

func requestIDFrom(r *http.Request) string {
	if value := strings.TrimSpace(r.Header.Get("X-Request-ID")); value != "" {
		return value
	}

	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(data[:])
}

func RequestID(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey{}).(string)
	return value
}
