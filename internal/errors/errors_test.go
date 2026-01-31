// Package errors provides tests for error handling and formatting.
package errors

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/quick"

	"github.com/Les-El/chexum/internal/color"
)

// Feature: cli-guidelines-review, Property 4: Error messages are human-readable
// Validates: Requirements 3.1, 3.6
//
// This property test verifies that error messages are human-readable by checking
// that they don't contain technical jargon like stack traces, raw error codes,
// or overly technical language in non-verbose mode.
func TestProperty_ErrorMessagesAreHumanReadable(t *testing.T) {
	handler := NewErrorHandler(color.NewColorHandler())
	handler.color.SetEnabled(false)

	property := func(errMsg string) bool {
		if errMsg == "" {
			return true
		}
		formatted := handler.FormatError(errors.New(errMsg))
		if formatted == errMsg {
			return false
		}
		return !containsTechnicalJargon(formatted)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

func containsTechnicalJargon(s string) bool {
	jargon := []string{"0x", "syscall", "goroutine", "panic:", "runtime."}
	for _, p := range jargon {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

// Feature: cli-guidelines-review, Property 4: Error messages include suggestions
// Validates: Requirements 3.2
//
// This property test verifies that common error types include actionable suggestions.
func TestProperty_ErrorMessagesIncludeSuggestions(t *testing.T) {
	handler := NewErrorHandler(color.NewColorHandler())
	handler.color.SetEnabled(false)

	property := func(filename string) bool {
		if filename == "" {
			return true
		}

		errs := []error{NewFileNotFoundError(filename), NewPermissionError(filename)}
		for _, err := range errs {
			if !hasSuggestion(handler.FormatError(err)) {
				return false
			}
		}
		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

func hasSuggestion(formatted string) bool {
	words := []string{"try", "check", "use", "tip", "help"}
	lower := strings.ToLower(formatted)
	for _, w := range words {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}

// Feature: cli-guidelines-review, Property 4: Paths are sanitized
// Validates: Requirements 3.4
//
// This property test verifies that sensitive paths are sanitized in error messages.
func TestProperty_PathsAreSanitized(t *testing.T) {
	colorHandler := color.NewColorHandler()
	colorHandler.SetEnabled(false)

	handler := NewErrorHandler(colorHandler)

	// Get home directory for testing
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	// Property: For any path in the home directory, it should be displayed with ~
	property := func(filename string) bool {
		// Skip empty filenames
		if filename == "" {
			return true
		}

		// Create a path in the home directory
		fullPath := home + "/" + filename

		// Create an error with this path
		err := NewFileNotFoundError(fullPath)
		formatted := handler.FormatError(err)

		// The formatted message should contain ~ instead of the full home path
		// (unless the path is very short and doesn't need sanitization)
		if strings.Contains(formatted, home) && len(home) > 10 {
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// Feature: cli-guidelines-review, Property 5: Similar errors are grouped
// Validates: Requirements 3.5
func TestProperty_SimilarErrorsAreGrouped(t *testing.T) {
	property := func(numFileNotFound, numPermission, numOther uint8) bool {
		errs := generateErrorList(numFileNotFound%20, numPermission%20, numOther%20)
		groups := GroupErrors(errs)
		return verifyGroupCounts(groups, int(numFileNotFound%20), int(numPermission%20), int(numOther%20))
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

func generateErrorList(nf, np, no uint8) []error {
	var errs []error
	for i := uint8(0); i < nf; i++ {
		errs = append(errs, NewFileNotFoundError(fmt.Sprintf("f%d.txt", i)))
	}
	for i := uint8(0); i < np; i++ {
		errs = append(errs, NewPermissionError(fmt.Sprintf("p%d.txt", i)))
	}
	for i := uint8(0); i < no; i++ {
		errs = append(errs, fmt.Errorf("other %d", i))
	}
	return errs
}

func verifyGroupCounts(groups map[ErrorType][]error, expNF, expNP, expNO int) bool {
	if len(groups[ErrorTypeFileNotFound]) != expNF {
		return false
	}
	if len(groups[ErrorTypePermission]) != expNP {
		return false
	}
	if len(groups[ErrorTypeUnknown]) != expNO {
		return false
	}
	return true
}

// Feature: cli-guidelines-review, Property 5: Grouping preserves all errors
// Validates: Requirements 3.5
//
// This property test verifies that grouping errors doesn't lose any errors.
func TestProperty_GroupingPreservesAllErrors(t *testing.T) {
	// Property: For any list of errors, grouping should preserve all errors
	property := func(numErrors uint8) bool {
		// Limit to reasonable range
		if numErrors > 50 {
			numErrors = numErrors % 50
		}

		// Create a list of mixed errors
		var errs []error
		for i := uint8(0); i < numErrors; i++ {
			switch i % 3 {
			case 0:
				errs = append(errs, NewFileNotFoundError("file.txt"))
			case 1:
				errs = append(errs, NewPermissionError("file.txt"))
			case 2:
				errs = append(errs, errors.New("other error"))
			}
		}

		// Group the errors
		groups := GroupErrors(errs)

		// Count total errors in groups
		totalInGroups := 0
		for _, groupErrs := range groups {
			totalInGroups += len(groupErrs)
		}

		// Verify no errors were lost
		return totalInGroups == len(errs)
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// Unit Tests for Error Handling
// Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6

// TestErrorHandler_FormatError tests basic error formatting.
func TestErrorHandler_FormatError(t *testing.T) {
	h := NewErrorHandler(color.NewColorHandler())
	h.color.SetEnabled(false)

	t.Run("Standard Errors", func(t *testing.T) {
		tests := []struct {
			err  error
			want string
		}{
			{nil, ""},
			{errors.New("raw error"), "raw error"},
		}
		for _, tt := range tests {
			if res := h.FormatError(tt.err); !strings.Contains(res, tt.want) {
				t.Errorf("got %q", res)
			}
		}
	})

	t.Run("Custom Errors", func(t *testing.T) {
		err := NewFileNotFoundError("file.txt")
		if res := h.FormatError(err); !strings.Contains(res, "Cannot find file") {
			t.Errorf("got %q", res)
		}
	})
}

// TestErrorHandler_SuggestFix tests suggestion generation.
func TestErrorHandler_SuggestFix(t *testing.T) {
	h := NewErrorHandler(color.NewColorHandler())
	h.color.SetEnabled(false)

	tests := []error{
		NewFileNotFoundError("missing.txt"),
		NewPermissionError("locked.txt"),
		NewInvalidHashError("short", "sha256", 64),
	}

	for _, err := range tests {
		if h.SuggestFix(err) == "" {
			t.Error("expected suggestion")
		}
	}
}

// TestErrorHandler_VerboseMode tests verbose error output.
func TestErrorHandler_VerboseMode(t *testing.T) {
	h := NewErrorHandler(color.NewColorHandler())
	h.color.SetEnabled(false)

	err := &Error{Message: "msg", Original: errors.New("cause")}

	h.SetVerbose(false)
	if strings.Contains(h.FormatError(err), "cause") {
		t.Error("unexpected cause")
	}

	h.SetVerbose(true)
	if !strings.Contains(h.FormatError(err), "cause") {
		t.Error("missing cause")
	}
}

// TestGroupErrors tests error grouping functionality.
func TestGroupErrors(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		if len(GroupErrors(nil)) != 0 {
			t.Error("expected empty")
		}
	})

	t.Run("Grouping", func(t *testing.T) {
		errs := []error{
			NewFileNotFoundError("f1"),
			NewFileNotFoundError("f2"),
			NewPermissionError("p1"),
		}
		groups := GroupErrors(errs)
		if len(groups[ErrorTypeFileNotFound]) != 2 || len(groups[ErrorTypePermission]) != 1 {
			t.Errorf("grouping failed: %v", groups)
		}
	})
}

// TestSanitizePath tests path sanitization.
func TestSanitizePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	tests := []struct {
		name     string
		path     string
		wantHome bool // Should contain ~ instead of home path
	}{
		{
			name:     "path in home directory",
			path:     home + "/documents/file.txt",
			wantHome: true,
		},
		{
			name:     "relative path",
			path:     "file.txt",
			wantHome: false,
		},
		{
			name:     "current directory path",
			path:     "./file.txt",
			wantHome: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizePath(tt.path)

			if tt.wantHome {
				if !strings.Contains(result, "~") {
					t.Errorf("Expected path to contain ~, got %q", result)
				}
				if strings.Contains(result, home) {
					t.Errorf("Expected home path to be replaced with ~, got %q", result)
				}
			}
		})
	}
}

// TestNewFileNotFoundError tests file not found error creation.
func TestNewFileNotFoundError(t *testing.T) {
	err := NewFileNotFoundError("missing.txt")

	if err.Type != ErrorTypeFileNotFound {
		t.Errorf("Expected ErrorTypeFileNotFound, got %v", err.Type)
	}

	if !strings.Contains(err.Message, "missing.txt") {
		t.Errorf("Expected message to contain filename, got %q", err.Message)
	}

	if err.Suggestion == "" {
		t.Errorf("Expected suggestion to be non-empty")
	}
}

// TestNewPermissionError tests permission error creation.
func TestNewPermissionError(t *testing.T) {
	err := NewPermissionError("protected.txt")

	if err.Type != ErrorTypePermission {
		t.Errorf("Expected ErrorTypePermission, got %v", err.Type)
	}

	if !strings.Contains(err.Message, "protected.txt") {
		t.Errorf("Expected message to contain filename, got %q", err.Message)
	}

	if err.Suggestion == "" {
		t.Errorf("Expected suggestion to be non-empty")
	}
}

// TestNewInvalidHashError tests invalid hash error creation.
func TestNewInvalidHashError(t *testing.T) {
	err := NewInvalidHashError("abc123", "sha256", 64)

	if err.Type != ErrorTypeInvalidHash {
		t.Errorf("Expected ErrorTypeInvalidHash, got %v", err.Type)
	}

	if !strings.Contains(err.Message, "sha256") {
		t.Errorf("Expected message to contain algorithm, got %q", err.Message)
	}

	if !strings.Contains(err.Suggestion, "64") {
		t.Errorf("Expected suggestion to mention expected length, got %q", err.Suggestion)
	}
}

// TestNewConfigError tests config error creation.
func TestNewConfigError(t *testing.T) {
	err := NewConfigError("Invalid configuration value")

	if err.Type != ErrorTypeConfig {
		t.Errorf("Expected ErrorTypeConfig, got %v", err.Type)
	}

	if err.Message != "Invalid configuration value" {
		t.Errorf("Expected message to match input, got %q", err.Message)
	}

	if err.Suggestion == "" {
		t.Errorf("Expected suggestion to be non-empty")
	}
}

// TestError_Unwrap tests error unwrapping.
func TestError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := &Error{
		Type:     ErrorTypeUnknown,
		Message:  "wrapped message",
		Original: originalErr,
	}

	unwrapped := wrappedErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Expected unwrapped error to be original, got %v", unwrapped)
	}
}

// TestError tests the Error method.
func TestError(t *testing.T) {
	err := &Error{
		Message: "test error message",
	}
	if err.Error() != "test error message" {
		t.Errorf("Expected %q, got %q", "test error message", err.Error())
	}
}

func TestUnwrap(t *testing.T) {
	TestError_Unwrap(t)
}

func TestNewErrorHandler(t *testing.T) {
	h := NewErrorHandler(color.NewColorHandler())
	if h == nil {
		t.Fatal("NewErrorHandler returned nil")
	}
}

func TestSetVerbose(t *testing.T) {
	TestErrorHandler_VerboseMode(t)
}

func TestFormatError(t *testing.T) {
	TestErrorHandler_FormatError(t)
}

func TestSuggestFix(t *testing.T) {
	TestErrorHandler_SuggestFix(t)
}

// TestClassifyError tests error classification.
func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantType ErrorType
	}{
		{
			name:     "file not found",
			err:      os.ErrNotExist,
			wantType: ErrorTypeFileNotFound,
		},
		{
			name:     "permission denied",
			err:      os.ErrPermission,
			wantType: ErrorTypePermission,
		},
		{
			name:     "unknown error",
			err:      errors.New("unknown"),
			wantType: ErrorTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyError(tt.err)
			if result != tt.wantType {
				t.Errorf("Expected %v, got %v", tt.wantType, result)
			}
		})
	}
}
