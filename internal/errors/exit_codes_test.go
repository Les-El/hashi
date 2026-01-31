package errors

import (
	"fmt"
	"os"
	"testing"
	"testing/quick"

	"github.com/Les-El/chexum/internal/config"
	"github.com/Les-El/chexum/internal/hash"
)

func TestDetermineExitCode(t *testing.T) {
	cfg := config.DefaultConfig()

	t.Run("Success", func(t *testing.T) {
		res := &hash.Result{Entries: []hash.Entry{{Hash: "a"}}, Matches: []hash.MatchGroup{{Hash: "a", Count: 1}}}
		if c := DetermineExitCode(cfg, res); c != config.ExitSuccess {
			t.Errorf("got %d", c)
		}
	})

	t.Run("Match Required", func(t *testing.T) {
		cMatch := config.DefaultConfig()
		cMatch.MatchRequired = true
		res := &hash.Result{Unmatched: []hash.Entry{{Hash: "a"}}}
		if c := DetermineExitCode(cMatch, res); c != config.ExitNoMatches {
			t.Errorf("got %d", c)
		}
	})

	t.Run("Errors", func(t *testing.T) {
		res := &hash.Result{Errors: []error{fmt.Errorf("err")}}
		if c := DetermineExitCode(cfg, res); c != config.ExitPartialFailure {
			t.Errorf("got %d", c)
		}
	})
}

func TestDetermineDiscoveryExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"not exist", os.ErrNotExist, config.ExitFileNotFound},
		{"permission", os.ErrPermission, config.ExitPermissionErr},
		{"other", fmt.Errorf("other"), config.ExitPartialFailure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetermineDiscoveryExitCode(tt.err); got != tt.want {
				t.Errorf("DetermineDiscoveryExitCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestProperty_ExitCodes verifies universal exit code properties.
func TestProperty_ExitCodes(t *testing.T) {
	// Property 16: Exit codes reflect processing status
	f := func(matchRequired bool, hasMatches bool, hasErrors bool) bool {
		cfg := config.DefaultConfig()
		cfg.MatchRequired = matchRequired

		result := &hash.Result{}
		if hasMatches {
			result.Matches = []hash.MatchGroup{{Hash: "abc", Count: 1}}
		}
		if hasErrors {
			result.Errors = []error{fmt.Errorf("error")}
			result.Entries = []hash.Entry{{Error: fmt.Errorf("error")}}
		} else {
			result.Entries = []hash.Entry{{Hash: "abc"}}
		}

		code := DetermineExitCode(cfg, result)

		if hasErrors {
			return code != config.ExitSuccess
		}
		if matchRequired && !hasMatches {
			return code == config.ExitNoMatches
		}
		return code == config.ExitSuccess
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
