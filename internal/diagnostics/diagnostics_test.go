package diagnostics

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Les-El/chexum/internal/config"
	"github.com/Les-El/chexum/internal/console"
)

func TestRunDiagnostics(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "chexum-diag-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		cfg      *config.Config
		contains []string
	}{
		{"Basic info", &config.Config{Algorithm: "sha256"}, []string{"System Information", "Algorithm 'sha256' sanity check passed"}},
		{"File inspection", &config.Config{Algorithm: "sha256", Files: []string{testFile}}, []string{"Inspecting 1 input arguments", "Checking '" + testFile + "'", "Exists: YES", "Size: 5 bytes", "Readable: YES"}},
		{"Missing file", &config.Config{Algorithm: "sha256", Files: []string{filepath.Join(tmpDir, "missing.txt")}}, []string{"Exists: NO", "Parent directory exists: YES"}},
		{"Hash inspection", &config.Config{Algorithm: "sha256", Hashes: []string{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"}}, []string{"Inspecting 1 hash arguments", "Valid format: YES", "Possible algorithms: sha256"}},
		{"Invalid hash inspection", &config.Config{Algorithm: "sha256", Hashes: []string{"not-a-hash"}}, []string{"Valid format: NO", "Length: 10"}},
		{"Algorithm failure", &config.Config{Algorithm: "invalid-algo"}, []string{"Algorithm 'invalid-algo' check FAILED"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf bytes.Buffer
			streams := &console.Streams{Out: &outBuf, Err: &outBuf}
			exitCode := RunDiagnostics(tt.cfg, streams)
			if exitCode != config.ExitSuccess {
				t.Errorf("Expected exit code %d, got %d", config.ExitSuccess, exitCode)
			}
			output := outBuf.String()
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("Output missing expected string: %q", s)
				}
			}
		})
	}
}
