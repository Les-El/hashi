// Package security provides input validation and path safety checks for chexum.
//
// DESIGN PRINCIPLE: Chexum can't change Chexum
// -----------------------------------------
// One of the most dangerous vulnerabilities in automated tools is
// "Self-Modification" or "Configuration Injection". If an attacker can force
// a tool to overwrite its own security policy, the tool becomes a weapon.
//
// Chexum defends against this with two core mandates:
//  1. READ-ONLY ON SOURCE: Chexum never, under any circumstances, modifies the
//     files it is hashing.
//  2. PROTECTED CONFIGURATION: Chexum cannot write output or logs to its own
//     configuration files or directories.
//
// This package implements these mandates through strict path validation,
// extension whitelisting, and obfuscated error messages that prevent
// attackers from discovering which files are protected.
package security

import (
	"fmt"
	"os"
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
	".env",
	".chexum.toml",
	"*.key",
	"*.pem",
	"id_rsa",
	"id_ed25519",
}

var DefaultBlacklistDirs = []string{
	".git",
	".ssh",
	".aws",
	".config/chexum",
}

// ValidateOutputPath ensures an output path is safe to write to.
func ValidateOutputPath(path string, opts Options) error {
	if path == "" {
		return nil
	}

	if err := checkExtension(path); err != nil {
		return err
	}

	if err := checkTraversal(path); err != nil {
		return err
	}

	if err := ValidateFileName(filepath.Base(path), opts); err != nil {
		return err
	}

	if err := ValidateDirPath(path, opts); err != nil {
		return err
	}

	return checkFileStatus(path, opts)
}

func checkExtension(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	allowedExts := []string{".txt", ".json", ".jsonl", ".csv", ".log"}
	for _, allowed := range allowedExts {
		if ext == allowed {
			return nil
		}
	}
	return fmt.Errorf("output files must have extension: %s (got %s)",
		strings.Join(allowedExts, ", "), ext)
}

func checkTraversal(path string) error {
	cleaned := filepath.Clean(path)
	if strings.Contains(filepath.ToSlash(cleaned), "../") || filepath.Base(cleaned) == ".." {
		return fmt.Errorf("directory traversal not allowed in output path")
	}
	return nil
}

func checkFileStatus(path string, opts Options) error {
	info, err := os.Lstat(path)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return formatSecurityError(opts.Verbose, "cannot write to symlink")
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check file status: %w", err)
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
func (opts Options) isWhitelistedDir(part string) bool {
	partLower := strings.ToLower(part)
	for _, white := range opts.WhitelistDirs {
		whiteLower := strings.ToLower(white)
		if wMatched, _ := filepath.Match(whiteLower, partLower); wMatched {
			return true
		}
	}
	return false
}

// ValidateDirPath checks if any part of the path matches blacklisted directory names.
func ValidateDirPath(path string, opts Options) error {
	if path == "" {
		return nil
	}

	if err := checkGeneralTraversal(path, opts.Verbose); err != nil {
		return err
	}

	if err := checkConfigProtection(path, opts.Verbose); err != nil {
		return err
	}

	return checkBlacklistedDirs(path, opts)
}

func checkGeneralTraversal(path string, verbose bool) error {
	cleaned := filepath.Clean(path)
	if strings.Contains(filepath.ToSlash(cleaned), "../") || filepath.Base(cleaned) == ".." {
		return formatSecurityError(verbose, "directory traversal not allowed in paths")
	}
	return nil
}

func checkConfigProtection(path string, verbose bool) error {
	pathLower := strings.ToLower(path)
	if strings.Contains(pathLower, ".chexum") || strings.Contains(pathLower, ".config/chexum") {
		return formatSecurityError(verbose, "cannot write to configuration directory")
	}
	return nil
}

func checkBlacklistedDirs(path string, opts Options) error {
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
		if opts.isWhitelistedDir(part) {
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
				return formatSecurityError(opts.Verbose, fmt.Sprintf("cannot access directory matching security pattern: %s", pattern))
			}
		}
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
	// Use Clean to checking for traversal
	cleaned := filepath.Clean(path)
	if strings.Contains(filepath.ToSlash(cleaned), "../") || filepath.Base(cleaned) == ".." {
		return "", fmt.Errorf("directory traversal not allowed in paths")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}
	return abs, nil
}

// SanitizeOutput neutralizes non-printable characters and terminal escape sequences.
// This prevents "Terminal Escape Injection" where a malicious filename could
// execute commands or hide output in a user's terminal emulator.
func SanitizeOutput(s string) string {
	var sb strings.Builder
	for _, r := range s {
		// Replace control characters (0-31 and 127) with a placeholder
		// We allow common whitespace like space, but block \r, \n, \t in names
		// to prevent multi-line spoofing or alignment tricks.
		if r < 32 || r == 127 {
			sb.WriteRune('?')
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func formatSecurityError(verbose bool, details string) error {
	if verbose {
		return fmt.Errorf("security policy violation: %s", details)
	}
	return fmt.Errorf("Unknown write/append error") // Obfuscated for security
}
