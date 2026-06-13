package config

import (
	"fmt"
	"sync"
	"time"
)

type Snapshot struct {
	Ready            bool          `json:"ready"`
	ErrorRatePercent float64       `json:"error_rate_percent"`
	Latency          time.Duration `json:"-"`
	LatencyText      string        `json:"latency"`
	LatencyMillis    int64         `json:"latency_ms"`
}

type State struct {
	mu               sync.RWMutex
	ready            bool
	errorRatePercent float64
	latency          time.Duration
}

func NewState() *State {
	return &State{
		ready: true,
	}
}

func (s *State) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Snapshot{
		Ready:            s.ready,
		ErrorRatePercent: s.errorRatePercent,
		Latency:          s.latency,
		LatencyText:      s.latency.String(),
		LatencyMillis:    s.latency.Milliseconds(),
	}
}

func (s *State) SetReady(ready bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ready = ready
}

func (s *State) SetErrorRatePercent(value float64) error {
	if value < 0 || value > 100 {
		return fmt.Errorf("error_rate_percent must be between 0 and 100")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.errorRatePercent = value
	return nil
}

func (s *State) SetLatency(value time.Duration) error {
	if value < 0 {
		return fmt.Errorf("latency must be greater than or equal to 0")
	}
	if value > 30*time.Second {
		return fmt.Errorf("latency must not exceed 30s")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.latency = value
	return nil
}
