// Package errors provides error handling and formatting for chexum.
//
// It transforms technical errors into user-friendly messages with
// actionable suggestions. Errors are grouped by type to reduce noise,
// and sensitive paths are sanitized in error messages.
package errors

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Les-El/chexum/internal/color"
	"github.com/Les-El/chexum/internal/security"
)

// ErrorType categorizes errors for grouping and handling.
type ErrorType int

const (
	// ErrorTypeUnknown is an unclassified error
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeFileNotFound indicates a file was not found
	ErrorTypeFileNotFound
	// ErrorTypePermission indicates a permission error
	ErrorTypePermission
	// ErrorTypeInvalidInput indicates invalid user input
	ErrorTypeInvalidInput
	// ErrorTypeInvalidHash indicates an invalid hash format
	ErrorTypeInvalidHash
	// ErrorTypeConfig indicates a configuration error
	ErrorTypeConfig
)

// Error wraps an error with additional context for user-friendly display.
type Error struct {
	Type       ErrorType
	Message    string
	Suggestion string
	Original   error
	Path       string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Original
}

// Handler formats errors for user-friendly display.
type Handler struct {
	color   *color.Handler
	verbose bool
}

// NewErrorHandler creates a new error handler.
func NewErrorHandler(colorHandler *color.Handler) *Handler {
	return &Handler{
		color:   colorHandler,
		verbose: false,
	}
}

// SetVerbose enables or disables verbose error output.
func (h *Handler) SetVerbose(verbose bool) {
	h.verbose = verbose
}

// FormatError formats an error for display.
func (h *Handler) FormatError(err error) string {
	if err == nil {
		return ""
	}

	// Check if it's our custom error type
	var chexumErr *Error
	if errors.As(err, &chexumErr) {
		return h.formatChexumError(chexumErr)
	}

	// Handle standard errors
	return h.formatStandardError(err)
}

// formatChexumError formats a chexum-specific error.
func (h *Handler) formatChexumError(err *Error) string {
	var sb strings.Builder

	// Error message with icon
	sb.WriteString(h.color.Error(err.Message))
	sb.WriteString("\n")

	// Suggestion if available
	if err.Suggestion != "" {
		sb.WriteString("\n  ")
		sb.WriteString(err.Suggestion)
		sb.WriteString("\n")
	}

	// Original error in verbose mode
	if h.verbose && err.Original != nil {
		sb.WriteString("\n  ")
		sb.WriteString(h.color.Gray(fmt.Sprintf("Details: %v", err.Original)))
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatStandardError formats a standard Go error.
func (h *Handler) formatStandardError(err error) string {
	// Classify the error
	errType := classifyError(err)

	// Create a user-friendly message
	message, suggestion := h.createMessage(err, errType)

	var sb strings.Builder
	sb.WriteString(h.color.Error(message))
	sb.WriteString("\n")

	if suggestion != "" {
		sb.WriteString("\n  ")
		sb.WriteString(suggestion)
		sb.WriteString("\n")
	}

	return sb.String()
}

// classifyError determines the type of a standard error.
func classifyError(err error) ErrorType {
	if os.IsNotExist(err) {
		return ErrorTypeFileNotFound
	}
	if os.IsPermission(err) {
		return ErrorTypePermission
	}
	return ErrorTypeUnknown
}

// createMessage creates a user-friendly message for an error.
func (h *Handler) createMessage(err error, errType ErrorType) (message, suggestion string) {
	errStr := err.Error()

	switch errType {
	case ErrorTypeFileNotFound:
		// Extract path from error message
		path := extractPath(errStr)
		message = fmt.Sprintf("Cannot find file: %s", sanitizePath(path))
		suggestion = "Check the file name for typos, or use 'ls' to see available files."

	case ErrorTypePermission:
		path := extractPath(errStr)
		message = fmt.Sprintf("Cannot read file: %s", sanitizePath(path))
		suggestion = "Check file permissions, or try running with elevated privileges."

	default:
		// Sanitize the error message
		message = sanitizeErrorMessage(errStr)
		suggestion = "If this error persists, please report it with the --verbose flag output."
	}

	return message, suggestion
}

// SuggestFix returns a suggestion for fixing an error.
func (h *Handler) SuggestFix(err error) string {
	var chexumErr *Error
	if errors.As(err, &chexumErr) {
		return chexumErr.Suggestion
	}

	errType := classifyError(err)
	_, suggestion := h.createMessage(err, errType)
	return suggestion
}

// NewFileNotFoundError creates a file not found error.
func NewFileNotFoundError(path string) *Error {
	return &Error{
		Type:       ErrorTypeFileNotFound,
		Message:    fmt.Sprintf("Cannot find file: %s", sanitizePath(path)),
		Suggestion: "Check the file name for typos, or use 'ls' to see available files.",
		Path:       path,
	}
}

// NewPermissionError creates a permission error.
func NewPermissionError(path string) *Error {
	return &Error{
		Type:       ErrorTypePermission,
		Message:    fmt.Sprintf("Cannot read file: %s", sanitizePath(path)),
		Suggestion: "Check file permissions, or try running with elevated privileges.",
		Path:       path,
	}
}

// NewInvalidHashError creates an invalid hash format error.
func NewInvalidHashError(hash, algorithm string, expectedLen int) *Error {
	return &Error{
		Type:    ErrorTypeInvalidHash,
		Message: fmt.Sprintf("Invalid %s hash: %s", algorithm, truncateHash(hash)),
		Suggestion: fmt.Sprintf("%s hashes must be exactly %d hexadecimal characters.\n"+
			"  Your input has %d characters.", strings.ToUpper(algorithm), expectedLen, len(hash)),
	}
}

// NewConfigError creates a configuration error.
func NewConfigError(message string) *Error {
	return &Error{
		Type:       ErrorTypeConfig,
		Message:    message,
		Suggestion: "Check your configuration file or command-line flags.",
	}
}

// GroupErrors groups similar errors together.
func GroupErrors(errs []error) map[ErrorType][]error {
	groups := make(map[ErrorType][]error)

	for _, err := range errs {
		var chexumErr *Error
		var errType ErrorType

		if errors.As(err, &chexumErr) {
			errType = chexumErr.Type
		} else {
			errType = classifyError(err)
		}

		groups[errType] = append(groups[errType], err)
	}

	return groups
}

// sanitizePath removes sensitive information from file paths.
func sanitizePath(path string) string {
	// Get the base name and parent directory
	base := filepath.Base(path)
	dir := filepath.Dir(path)

	// If the path is in the home directory, replace with ~
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(dir, home) {
		dir = "~" + strings.TrimPrefix(dir, home)
	}

	// Return a reasonable path representation
	result := base
	if dir != "." {
		result = filepath.Join(dir, base)
	}

	return security.SanitizeOutput(result)
}

// sanitizeErrorMessage removes technical details from error messages.
func sanitizeErrorMessage(msg string) string {
	// Remove common technical prefixes
	prefixes := []string{
		"open ",
		"read ",
		"stat ",
		"lstat ",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(msg, prefix) {
			msg = strings.TrimPrefix(msg, prefix)
		}
	}

	return msg
}

// extractPath extracts a file path from an error message.
func extractPath(errStr string) string {
	// Common patterns: "open /path/to/file: ..." or "/path/to/file: ..."
	parts := strings.SplitN(errStr, ":", 2)
	if len(parts) > 0 {
		path := strings.TrimSpace(parts[0])
		path = strings.TrimPrefix(path, "open ")
		path = strings.TrimPrefix(path, "read ")
		path = strings.TrimPrefix(path, "stat ")
		path = strings.TrimPrefix(path, "lstat ")
		return path
	}
	return errStr
}

// truncateHash truncates a hash for display in error messages.
func truncateHash(hash string) string {
	if len(hash) > 20 {
		return hash[:10] + "..." + hash[len(hash)-10:]
	}
	return hash
}
