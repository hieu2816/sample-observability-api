package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sample-observability-api/internal/buildinfo"
	"sample-observability-api/internal/config"
	"sample-observability-api/internal/observability"
)

func TestHealth(t *testing.T) {
	handler, _ := testHandler()

	response := request(handler, http.MethodGet, "/health", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected health response, got %s", response.Body.String())
	}
}

func TestReadinessReflectsRuntimeState(t *testing.T) {
	handler, state := testHandler()

	state.SetReady(false)
	response := request(handler, http.MethodGet, "/ready", "")

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", response.Code)
	}
}

func TestSetReadiness(t *testing.T) {
	handler, _ := testHandler()

	response := request(handler, http.MethodPost, "/config/readiness", `{"ready":false}`)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", response.Code, response.Body.String())
	}

	response = request(handler, http.MethodGet, "/ready", "")
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 after config change, got %d", response.Code)
	}
}

func TestSetErrorRateValidation(t *testing.T) {
	handler, _ := testHandler()

	response := request(handler, http.MethodPost, "/config/error-rate", `{"error_rate_percent":150}`)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestVersion(t *testing.T) {
	handler, _ := testHandler()

	response := request(handler, http.MethodGet, "/version", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"version":"test"`) {
		t.Fatalf("expected test version, got %s", response.Body.String())
	}
}

func TestOrdersAndPayments(t *testing.T) {
	handler, _ := testHandler()

	orders := request(handler, http.MethodGet, "/api/orders", "")
	if orders.Code != http.StatusOK || !strings.Contains(orders.Body.String(), "ord-1001") {
		t.Fatalf("unexpected orders response code=%d body=%s", orders.Code, orders.Body.String())
	}

	payments := request(handler, http.MethodGet, "/api/payments", "")
	if payments.Code != http.StatusOK || !strings.Contains(payments.Body.String(), "pay-9001") {
		t.Fatalf("unexpected payments response code=%d body=%s", payments.Code, payments.Body.String())
	}
}

func TestSimulateError(t *testing.T) {
	handler, _ := testHandler()

	response := request(handler, http.MethodGet, "/simulate/error", "")

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", response.Code)
	}
}

func TestMetrics(t *testing.T) {
	handler, _ := testHandler()

	_ = request(handler, http.MethodGet, "/health", "")
	response := request(handler, http.MethodGet, "/metrics", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "sample_api_http_requests_total") {
		t.Fatalf("expected metrics output, got %s", response.Body.String())
	}
}

func TestMetricsRejectsPost(t *testing.T) {
	handler, _ := testHandler()

	response := request(handler, http.MethodPost, "/metrics", "")

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", response.Code)
	}
}

func TestSimulateMemoryStressDuration(t *testing.T) {
	handler, _ := testHandler()

	response := request(handler, http.MethodGet, "/simulate/memory-stress?mb=1&duration=0s", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"duration":"0s"`) {
		t.Fatalf("expected duration in response, got %s", response.Body.String())
	}
}

func testHandler() (http.Handler, *config.State) {
	state := config.NewState()
	server := NewServer(ServerOptions{
		State:   state,
		Metrics: observability.NewRecorder(),
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		Build: buildinfo.Info{
			Version:   "test",
			Commit:    "test",
			BuildTime: "test",
		},
	})
	return server.Handler(), state
}

func request(handler http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}
