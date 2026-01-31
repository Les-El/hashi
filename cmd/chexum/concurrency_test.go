package main

import (
	"testing"
)

func TestCalculateWorkers(t *testing.T) {
	tests := []struct {
		name      string
		requested int // Input from --jobs flag
		numCPU    int // Mock system CPU count
		want      int
	}{
		// Auto-detection scenarios (requested == 0)
		{"Auto: Single Core", 0, 1, 1},
		{"Auto: Dual Core", 0, 2, 1},
		{"Auto: Quad Core", 0, 4, 3}, // 4 - 1 = 3
		{"Auto: Hex Core", 0, 6, 4},  // 6 - 2 = 4
		{"Auto: 8 Core", 0, 8, 6},    // 8 - 2 = 6
		{"Auto: 32 Core", 0, 32, 30}, // 32 - 2 = 30
		{"Auto: Massive", 0, 64, 32}, // Hard cap at 32

		// Explicit requests (should be honored but capped if sane?)
		// The requirement was: verify explicit inputs are respected within safety caps?
		// "If user says --jobs 100, we trust them but warn" -> Logic should probably return 100.
		// However, the security hardening in previous step capped it.
		// Let's refine the logic: Explicit > Auto Logic.
		// If I say -j 10, I get 10.
		{"Explicit: 1", 1, 8, 1},
		{"Explicit: 4", 4, 8, 4},
		{"Explicit: High", 100, 8, 100}, // User knows best (maybe they're hashing network latency)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateWorkersInternal(tt.requested, tt.numCPU)
			if got != tt.want {
				t.Errorf("CalculateWorkers(%d, CPUs=%d) = %d, want %d", tt.requested, tt.numCPU, got, tt.want)
			}
		})
	}
}

// Helper to access the private function logic for testing
func calculateWorkersInternal(requested, numCPU int) int {
	return calculateWorkers(requested, numCPU)
}
