package observability

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"sample-observability-api/internal/buildinfo"
	"sample-observability-api/internal/config"
)

var durationBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

type Recorder struct {
	mu       sync.Mutex
	requests map[string]uint64
	errors   map[string]uint64
	buckets  map[string][]uint64
	sums     map[string]float64
	counts   map[string]uint64
}

func NewRecorder() *Recorder {
	return &Recorder{
		requests: make(map[string]uint64),
		errors:   make(map[string]uint64),
		buckets:  make(map[string][]uint64),
		sums:     make(map[string]float64),
		counts:   make(map[string]uint64),
	}
}

func (r *Recorder) Observe(method string, route string, status int, duration time.Duration) {
	statusText := fmt.Sprintf("%d", status)
	requestKey := labels(method, route, statusText)
	durationKey := labels(method, route, "")

	r.mu.Lock()
	defer r.mu.Unlock()

	r.requests[requestKey]++
	if status >= 500 {
		r.errors[requestKey]++
	}

	seconds := duration.Seconds()
	if _, ok := r.buckets[durationKey]; !ok {
		r.buckets[durationKey] = make([]uint64, len(durationBuckets))
	}
	for i, bucket := range durationBuckets {
		if seconds <= bucket {
			r.buckets[durationKey][i]++
		}
	}
	r.sums[durationKey] += seconds
	r.counts[durationKey]++
}

func (r *Recorder) Handler(state *config.State, build buildinfo.Info) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.Render(state.Snapshot(), build)))
	})
}

func (r *Recorder) Render(snapshot config.Snapshot, build buildinfo.Info) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	var b strings.Builder
	b.WriteString("# HELP sample_api_http_requests_total Total HTTP requests.\n")
	b.WriteString("# TYPE sample_api_http_requests_total counter\n")
	for _, key := range sortedKeys(r.requests) {
		fmt.Fprintf(&b, "sample_api_http_requests_total{%s} %d\n", key, r.requests[key])
	}

	b.WriteString("# HELP sample_api_http_errors_total Total HTTP 5xx responses.\n")
	b.WriteString("# TYPE sample_api_http_errors_total counter\n")
	for _, key := range sortedKeys(r.errors) {
		fmt.Fprintf(&b, "sample_api_http_errors_total{%s} %d\n", key, r.errors[key])
	}

	b.WriteString("# HELP sample_api_http_request_duration_seconds HTTP request duration histogram.\n")
	b.WriteString("# TYPE sample_api_http_request_duration_seconds histogram\n")
	for _, key := range sortedKeys(r.buckets) {
		for i, bucket := range durationBuckets {
			fmt.Fprintf(&b, "sample_api_http_request_duration_seconds_bucket{%s,le=\"%g\"} %d\n", key, bucket, r.buckets[key][i])
		}
		fmt.Fprintf(&b, "sample_api_http_request_duration_seconds_bucket{%s,le=\"+Inf\"} %d\n", key, r.counts[key])
		fmt.Fprintf(&b, "sample_api_http_request_duration_seconds_sum{%s} %g\n", key, r.sums[key])
		fmt.Fprintf(&b, "sample_api_http_request_duration_seconds_count{%s} %d\n", key, r.counts[key])
	}

	ready := 0
	if snapshot.Ready {
		ready = 1
	}
	b.WriteString("# HELP sample_api_ready Current readiness state, 1 means ready.\n")
	b.WriteString("# TYPE sample_api_ready gauge\n")
	fmt.Fprintf(&b, "sample_api_ready %d\n", ready)

	b.WriteString("# HELP sample_api_configured_error_rate_percent Configured probabilistic error rate.\n")
	b.WriteString("# TYPE sample_api_configured_error_rate_percent gauge\n")
	fmt.Fprintf(&b, "sample_api_configured_error_rate_percent %g\n", snapshot.ErrorRatePercent)

	b.WriteString("# HELP sample_api_configured_latency_seconds Configured artificial latency in seconds.\n")
	b.WriteString("# TYPE sample_api_configured_latency_seconds gauge\n")
	fmt.Fprintf(&b, "sample_api_configured_latency_seconds %g\n", snapshot.Latency.Seconds())

	b.WriteString("# HELP sample_api_build_info Build metadata.\n")
	b.WriteString("# TYPE sample_api_build_info gauge\n")
	fmt.Fprintf(&b, "sample_api_build_info{version=\"%s\",commit=\"%s\",build_time=\"%s\"} 1\n", escape(build.Version), escape(build.Commit), escape(build.BuildTime))

	return b.String()
}

func labels(method string, route string, status string) string {
	parts := []string{
		fmt.Sprintf("method=\"%s\"", escape(method)),
		fmt.Sprintf("route=\"%s\"", escape(route)),
	}
	if status != "" {
		parts = append(parts, fmt.Sprintf("status=\"%s\"", escape(status)))
	}
	return strings.Join(parts, ",")
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func escape(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	return value
}
