package simulate

import (
	"runtime"
	"time"
)

func CPUStress(duration time.Duration) uint64 {
	deadline := time.Now().Add(duration)
	var iterations uint64
	var work uint64 = 1

	for {
		for i := 0; i < 10000; i++ {
			work = work*1664525 + 1013904223
			iterations++
		}

		if time.Now().After(deadline) {
			break
		}
	}

	_ = work
	return iterations
}

func MemoryStress(megabytes int, duration time.Duration) int {
	if megabytes < 1 {
		megabytes = 1
	}

	size := megabytes * 1024 * 1024
	data := make([]byte, size)
	for i := 0; i < len(data); i += 4096 {
		data[i] = byte(i % 255)
	}
	if duration > 0 {
		time.Sleep(duration)
	}
	runtime.KeepAlive(data)
	return len(data)
}
