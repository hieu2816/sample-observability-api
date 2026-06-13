package config

import (
	"testing"
	"time"
)

func TestStateDefaultsToReady(t *testing.T) {
	state := NewState()
	snapshot := state.Snapshot()

	if !snapshot.Ready {
		t.Fatal("expected default readiness to be true")
	}
	if snapshot.ErrorRatePercent != 0 {
		t.Fatalf("expected default error rate 0, got %v", snapshot.ErrorRatePercent)
	}
	if snapshot.Latency != 0 {
		t.Fatalf("expected default latency 0, got %v", snapshot.Latency)
	}
}

func TestSetErrorRateValidation(t *testing.T) {
	state := NewState()

	if err := state.SetErrorRatePercent(50); err != nil {
		t.Fatalf("expected valid error rate, got %v", err)
	}
	if err := state.SetErrorRatePercent(-1); err == nil {
		t.Fatal("expected negative error rate to fail")
	}
	if err := state.SetErrorRatePercent(101); err == nil {
		t.Fatal("expected error rate above 100 to fail")
	}
}

func TestSetLatencyValidation(t *testing.T) {
	state := NewState()

	if err := state.SetLatency(250 * time.Millisecond); err != nil {
		t.Fatalf("expected valid latency, got %v", err)
	}
	if err := state.SetLatency(-1 * time.Second); err == nil {
		t.Fatal("expected negative latency to fail")
	}
	if err := state.SetLatency(31 * time.Second); err == nil {
		t.Fatal("expected latency above 30s to fail")
	}
}
