package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/quick"
)

func TestValidateOutputPath(t *testing.T) {
	opts := Options{Verbose: true}

	t.Run("valid path", func(t *testing.T) {
		if err := ValidateOutputPath("results.txt", opts); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid extension", func(t *testing.T) {
		if err := ValidateOutputPath("results.exe", opts); err == nil {
			t.Error("expected error for .exe")
		}
	})

	t.Run("directory traversal", func(t *testing.T) {
		if err := ValidateOutputPath("../../etc/passwd.txt", opts); err == nil {
			t.Error("expected error for traversal")
		}
	})

	t.Run("blacklisted file", func(t *testing.T) {
		if err := ValidateOutputPath("id_rsa.txt", opts); err == nil {
			t.Error("expected error for 'id_rsa' file")
		}
	})

	t.Run("blacklisted dir", func(t *testing.T) {
		if err := ValidateOutputPath(".git/results.txt", opts); err == nil {
			t.Error("expected error for '.git' dir")
		}
	})

	t.Run("whitelist override", func(t *testing.T) {
		wOpts := opts
		wOpts.WhitelistFiles = []string{"secret_report.txt"}
		if err := ValidateOutputPath("secret_report.txt", wOpts); err != nil {
			t.Errorf("expected whitelist to allow file, got %v", err)
		}
	})
}

func TestValidateOutputPath_Symlink(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target.txt")
	os.WriteFile(target, []byte("data"), 0600)
	link := filepath.Join(tmpDir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("skipping symlink test: ", err)
	}

	if err := ValidateOutputPath(link, Options{Verbose: true}); err == nil {
		t.Error("expected error for symlink")
	} else if !strings.Contains(err.Error(), "symlink") {
		t.Errorf("expected symlink error, got %v", err)
	}
}

func TestProperty_SecurityValidation(t *testing.T) {
	// Property 8: Input validation occurs before processing
	// We verify that blacklisted patterns are always rejected unless whitelisted
	f := func(name string) bool {
		if name == "" {
			return true
		}
		opts := Options{Verbose: true}

		// If name contains a blacklist word, it should be rejected
		isBlacklisted := false
		for _, b := range DefaultBlacklistFiles {
			if strings.Contains(strings.ToLower(name), strings.ToLower(b)) {
				isBlacklisted = true
				break
			}
		}

		err := ValidateFileName(name, opts)
		if isBlacklisted && !strings.Contains(name, "*") && !strings.Contains(name, "?") {
			return err != nil
		}
		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestResolveSafePath(t *testing.T) {
	t.Run("safe path", func(t *testing.T) {
		_, err := ResolveSafePath("file.txt")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("traversal", func(t *testing.T) {
		_, err := ResolveSafePath("../outside")
		if err == nil {
			t.Error("expected error for ..")
		}
	})
}

func TestValidateFileName(t *testing.T) {
	opts := Options{Verbose: true}
	if err := ValidateFileName("id_rsa", opts); err == nil {
		t.Error("Expected error for id_rsa")
	}
	if err := ValidateFileName("safe.txt", opts); err != nil {
		t.Errorf("Unexpected error for safe.txt: %v", err)
	}
}

func TestValidateDirPath(t *testing.T) {
	opts := Options{Verbose: true}
	if err := ValidateDirPath(".git/file.txt", opts); err == nil {
		t.Error("Expected error for .git/file.txt")
	}
	if err := ValidateDirPath("safe/file.txt", opts); err != nil {
		t.Errorf("Unexpected error for safe/file.txt: %v", err)
	}
}

func TestValidateInputs(t *testing.T) {
	opts := Options{Verbose: true}
	files := []string{"safe.txt"}
	hashes := []string{"abc123"}
	if err := ValidateInputs(files, hashes, opts); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	badHashes := []string{"not-hex"}
	if err := ValidateInputs(files, badHashes, opts); err == nil {
		t.Error("Expected error for invalid hex hash")
	}
}

func TestSanitizeOutput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"safe", "safe"},
		{"with\nnewline", "with?newline"},
		{"with\rescape", "with?escape"},
		{"with\tescape", "with?escape"},
		{"\x1b[31mred\x1b[0m", "?[31mred?[0m"},
	}

	for _, tc := range tests {
		got := SanitizeOutput(tc.input)
		if got != tc.expected {
			t.Errorf("SanitizeOutput(%q) = %q; expected %q", tc.input, got, tc.expected)
		}
	}
}
