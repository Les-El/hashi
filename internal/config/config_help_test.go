package config

import (
	"strings"
	"testing"
)

func TestHelpText(t *testing.T) {
	help := HelpText()
	if len(help) == 0 {
		t.Error("HelpText() returned empty string")
	}
	sections := []string{"EXAMPLES", "USAGE", "FLAGS", "EXIT CODES"}
	for _, section := range sections {
		if !strings.Contains(help, section) {
			t.Errorf("HelpText() missing section: %s", section)
		}
	}
}

func TestVersionText(t *testing.T) {
	version := VersionText()
	if len(version) == 0 {
		t.Error("VersionText() returned empty string")
	}
	if !strings.Contains(version, "chexum") {
		t.Error("VersionText() should contain 'chexum'")
	}
}
