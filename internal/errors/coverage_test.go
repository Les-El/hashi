package errors

import (
	"errors"
	"github.com/Les-El/chexum/internal/color"
	"os"
	"strings"
	"testing"
)

func TestErrorHandler_FormatError_ChexumErrorVerbose(t *testing.T) {
	colorHandler := color.NewColorHandler()
	colorHandler.SetEnabled(false)
	h := NewErrorHandler(colorHandler)
	h.SetVerbose(true)

	err := &Error{
		Message:    "test msg",
		Suggestion: "test sug",
		Original:   errors.New("original"),
	}

	formatted := h.FormatError(err)
	if !strings.Contains(formatted, "Details: original") {
		t.Error("Expected original error details in verbose mode")
	}
}

func TestSuggestFix_StandardError(t *testing.T) {
	colorHandler := color.NewColorHandler()
	colorHandler.SetEnabled(false)
	h := NewErrorHandler(colorHandler)

	err := os.ErrNotExist
	suggestion := h.SuggestFix(err)
	if suggestion == "" {
		t.Error("Expected suggestion for standard error")
	}

	chexumErr := NewConfigError("msg")
	if h.SuggestFix(chexumErr) != chexumErr.Suggestion {
		t.Error("Expected suggestion from chexumErr")
	}
}

func TestGroupErrors_Mixed(t *testing.T) {
	errs := []error{
		os.ErrNotExist,
		NewConfigError("config error"),
		errors.New("random"),
	}
	groups := GroupErrors(errs)
	if len(groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(groups))
	}
}

func TestExtractPath_EdgeCases(t *testing.T) {
	if extractPath("") != "" {
		t.Error("Expected empty string")
	}
}

func TestSanitizeErrorMessage_EdgeCases(t *testing.T) {
	if sanitizeErrorMessage("random") != "random" {
		t.Error("Expected unchanged message")
	}
}
