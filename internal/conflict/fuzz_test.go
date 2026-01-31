// Package conflict_test provides fuzzing for flag combinations.
package conflict_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/Les-El/chexum/internal/conflict"
)

// TestFuzzFlagConflicts generates random flag combinations to ensure no panics.
// **Validates: Task 37.4 (Fuzzing Moonshot)**
func TestFuzzFlagConflicts(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	modes := []string{"--bool", ""}
	formats := []string{"--json", "--jsonl", "--plain", "--format=verbose", "--format=default", ""}
	verbosities := []string{"--quiet", "--verbose", "-q", "-v", ""}

	for i := 0; i < 1000; i++ {
		args := []string{}
		flagSet := make(map[string]bool)

		// Randomly pick one of each category
		m := modes[rand.Intn(len(modes))]
		f := formats[rand.Intn(len(formats))]
		v := verbosities[rand.Intn(len(verbosities))]

		if m != "" {
			args = append(args, m)
			flagSet["bool"] = true
		}
		if f != "" {
			args = append(args, f)
			if f == "--json" {
				flagSet["json"] = true
			} else if f == "--jsonl" {
				flagSet["jsonl"] = true
			} else if f == "--plain" {
				flagSet["plain"] = true
			}
		}
		if v != "" {
			args = append(args, v)
			if v == "--quiet" || v == "-q" {
				flagSet["quiet"] = true
			} else if v == "--verbose" || v == "-v" {
				flagSet["verbose"] = true
			}
		}

		// Execute resolution
		// Since we don't care about order here, just pass whatever
		_, _, err := conflict.ResolveState(flagSet, "default")
		if err != nil {
			t.Errorf("Fuzz failed with args %v: %v", args, err)
		}
	}
}
