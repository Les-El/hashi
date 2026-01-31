package config

import (
	"fmt"
	"strings"

	"github.com/Les-El/chexum/internal/security"
)

// WriteError returns a generic error message for security-sensitive write failures.
func WriteError() error {
	return fmt.Errorf("Unknown write/append error")
}

// WriteErrorWithVerbose returns either a generic or detailed error message.
func WriteErrorWithVerbose(verbose bool, verboseDetails string) error {
	if verbose {
		return fmt.Errorf("%s", verboseDetails)
	}
	return WriteError()
}

// FileSystemError returns a generic error for file system operations.
func FileSystemError(verbose bool, verboseDetails string) error {
	if verbose {
		return fmt.Errorf("%s", verboseDetails)
	}
	return WriteError()
}

// HandleFileWriteError processes file writing errors.
func HandleFileWriteError(err error, verbose bool, path string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	sanitizedPath := security.SanitizeOutput(path)

	if strings.Contains(errStr, "permission denied") ||
		strings.Contains(errStr, "access is denied") {
		return FileSystemError(verbose, fmt.Sprintf("permission denied writing to %s", sanitizedPath))
	}

	if strings.Contains(errStr, "no space left") ||
		strings.Contains(errStr, "disk full") {
		return FileSystemError(verbose, fmt.Sprintf("insufficient disk space for %s", sanitizedPath))
	}

	if strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "timeout") {
		return FileSystemError(verbose, fmt.Sprintf("network error writing to %s", sanitizedPath))
	}

	if strings.Contains(errStr, "file name too long") ||
		strings.Contains(errStr, "path too long") {
		return FileSystemError(verbose, fmt.Sprintf("path too long: %s", sanitizedPath))
	}

	return err
}
