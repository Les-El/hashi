// Package security provides input validation and path safety checks for hashi.
package security

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Security options for validation.
type Options struct {
	Verbose        bool
	BlacklistFiles []string
	BlacklistDirs  []string
	WhitelistFiles []string
	WhitelistDirs  []string
}

// Default blacklists
var DefaultBlacklistFiles = []string{
	"config",
	"secret",
	"key",
	"password",
	"credential",
	".env",
	".hashi.toml",
}

var DefaultBlacklistDirs = []string{
	"config",
	"secret",
	"key",
	"password",
	"credential",
	".git",
	".ssh",
}

// ValidateOutputPath ensures an output path is safe to write to.
func ValidateOutputPath(path string, opts Options) error {
	if path == "" {
		return nil
	}

	// 1. Extension validation
	ext := strings.ToLower(filepath.Ext(path))
	allowedExts := []string{".txt", ".json", ".csv"}
	extAllowed := false
	for _, allowed := range allowedExts {
		if ext == allowed {
			extAllowed = true
			break
		}
	}
	if !extAllowed {
		return fmt.Errorf("output files must have extension: %s (got %s)",
			strings.Join(allowedExts, ", "), ext)
	}

	// 2. Directory traversal check
	if strings.Contains(path, "..") {
		return fmt.Errorf("directory traversal not allowed in output path")
	}

	// 3. File name validation
	basename := filepath.Base(path)
	if err := ValidateFileName(basename, opts); err != nil {
		return err
	}

	// 4. Directory validation
	if err := ValidateDirPath(path, opts); err != nil {
		return err
	}

	return nil
}

// ValidateFileName checks if a filename matches any security patterns.
func ValidateFileName(filename string, opts Options) error {
	if filename == "" {
		return nil
	}

	allBlacklist := append(DefaultBlacklistFiles, opts.BlacklistFiles...)
	filenameLower := strings.ToLower(filename)

	for _, pattern := range allBlacklist {
		patternLower := strings.ToLower(pattern)
		matched, _ := filepath.Match(patternLower, filenameLower)
		if !matched {
			// Also check for prefix match for non-glob patterns
			if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
				matched = strings.HasPrefix(filenameLower, patternLower)
			}
		}

		if matched {
			// Check whitelist
			for _, white := range opts.WhitelistFiles {
				whiteLower := strings.ToLower(white)
				if wMatched, _ := filepath.Match(whiteLower, filenameLower); wMatched {
					return nil
				}
				if !strings.Contains(white, "*") && !strings.Contains(white, "?") {
					if strings.HasPrefix(filenameLower, whiteLower) {
						return nil
					}
				}
			}
			return formatSecurityError(opts.Verbose, fmt.Sprintf("cannot write to file matching security pattern: %s", pattern))
		}
	}
	return nil
}

// ValidateDirPath checks if any part of the path matches blacklisted directory names.
func ValidateDirPath(path string, opts Options) error {
	if path == "" {
		return nil
	}

	// 1. General directory traversal check for ALL paths
	if strings.Contains(path, "..") {
		return formatSecurityError(opts.Verbose, "directory traversal not allowed in paths")
	}

	// 2. Explicit protection for hashi configuration
	pathLower := strings.ToLower(path)
	if strings.Contains(pathLower, ".hashi") || strings.Contains(pathLower, ".config/hashi") {
		return formatSecurityError(opts.Verbose, "cannot write to configuration directory")
	}

	allBlacklist := append(DefaultBlacklistDirs, opts.BlacklistDirs...)
	dir := filepath.Dir(path)
	if dir == "." || dir == "/" {
		return nil
	}

	parts := strings.Split(filepath.Clean(dir), string(filepath.Separator))
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			continue
		}
		partLower := strings.ToLower(part)
		for _, pattern := range allBlacklist {
			patternLower := strings.ToLower(pattern)
			matched, _ := filepath.Match(patternLower, partLower)
			if !matched && !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
				matched = strings.HasPrefix(partLower, patternLower)
			}

			if matched {
				// Check whitelist
				for _, white := range opts.WhitelistDirs {
					whiteLower := strings.ToLower(white)
					if wMatched, _ := filepath.Match(whiteLower, partLower); wMatched {
						goto nextPart
					}
				}
				return formatSecurityError(opts.Verbose, fmt.Sprintf("cannot access directory matching security pattern: %s", pattern))
			}
		}
	nextPart:
	}
	return nil
}

// ValidateInputs performs security validation on all provided file paths and hash strings.
func ValidateInputs(files []string, hashes []string, opts Options) error {
	for _, file := range files {
		if file == "-" {
			continue
		}
		if err := ValidateDirPath(file, opts); err != nil {
			return err
		}
	}
	// Hashes are already classified and normalized, but we can do a sanity check
	for _, h := range hashes {
		if !isValidHex(h) {
			return fmt.Errorf("invalid hash string format: %s", h)
		}
	}
	return nil
}

func isValidHex(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// ResolveSafePath returns the absolute path while ensuring no traversal attempts.
func ResolveSafePath(path string) (string, error) {
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("directory traversal not allowed in paths")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}
	return abs, nil
}

func formatSecurityError(verbose bool, details string) error {
	if verbose {
		return fmt.Errorf("security policy violation: %s", details)
	}
	return fmt.Errorf("Unknown write/append error") // Obfuscated for security
}
