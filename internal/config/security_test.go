package config

import (
	"strings"
	"testing"

	"github.com/Les-El/chexum/internal/security"
)

func TestValidateOutputPath(t *testing.T) {
	t.Run("Allowed Extensions", testAllowedExtensions)
	t.Run("Blocked Extensions", testBlockedExtensions)
	t.Run("Blacklisted Paths", testBlacklistedPaths)
}

func testAllowedExtensions(t *testing.T) {
	paths := []string{"output.txt", "data.json", "data.jsonl", "report.csv", "output.log", "OUTPUT.TXT", "logs/out.txt"}
	cfg := DefaultConfig()
	for _, p := range paths {
		if err := validateOutputPath(p, cfg); err != nil {
			t.Errorf("%s should be allowed: %v", p, err)
		}
	}
}

func testBlockedExtensions(t *testing.T) {
	paths := []string{"script.sh", "malicious.py", "exe", "config.toml"}
	cfg := DefaultConfig()
	for _, p := range paths {
		if err := validateOutputPath(p, cfg); err == nil {
			t.Errorf("%s should be blocked", p)
		}
	}
}

func testBlacklistedPaths(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Verbose = true
	tests := []struct {
		path string
		msg  string
	}{
		{"id_rsa.txt", "security pattern"},
		{".git/config.txt", "security pattern"},
		{".chexum/out.txt", "configuration directory"},
	}
	for _, tt := range tests {
		err := validateOutputPath(tt.path, cfg)
		if err == nil || !strings.Contains(err.Error(), tt.msg) {
			t.Errorf("expected error for %s containing %q, got %v", tt.path, tt.msg, err)
		}
	}
}

var validateConfigSecurityTests = []struct {
	name       string
	outputFile string
	logFile    string
	logJSON    string
	shouldErr  bool
	errMsg     string
}{
	{
		name:       "all safe paths",
		outputFile: "results.txt",
		logFile:    "app.txt",
		logJSON:    "debug.json",
		shouldErr:  false,
	},
	{
		name:       "unsafe output file",
		outputFile: ".chexum.toml",
		shouldErr:  true,
		errMsg:     "output file",
	},
	{
		name:      "unsafe log file - default blacklist",
		logFile:   "id_rsa.txt",
		shouldErr: true,
		errMsg:    "log file",
	},
	{
		name:      "unsafe JSON log file",
		logJSON:   ".chexum/debug.json",
		shouldErr: true,
		errMsg:    "JSON log file",
	},
	{
		name:       "empty paths allowed",
		outputFile: "",
		logFile:    "",
		logJSON:    "",
		shouldErr:  false,
	},
}

func TestValidateConfigWithSecurity(t *testing.T) {
	for _, tt := range validateConfigSecurityTests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.OutputFile = tt.outputFile
			cfg.LogFile = tt.logFile
			cfg.LogJSON = tt.logJSON

			_, err := ValidateConfig(cfg)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("ValidateConfig() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateConfig() error %q should contain %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateConfig() expected no error, got %v", err)
				}
			}
		})
	}
}

// TestSecurityValidationFunctions tests the new security validation functions

func TestSecurityValidationFunctions(t *testing.T) {

	t.Run("FileName", testValidateFileName)

	t.Run("FileNameCustom", testValidateFileNameCustom)

	t.Run("DirPath", testValidateDirPath)

}

func testValidateFileName(t *testing.T) {

	opts := security.Options{Verbose: true}

	tests := []struct {
		filename string

		wantErr bool
	}{

		{"safe.txt", false},

		{".env", true},

		{".chexum.toml", true},
	}

	for _, tt := range tests {

		err := security.ValidateFileName(tt.filename, opts)

		if (err != nil) != tt.wantErr {

			t.Errorf("ValidateFileName(%s) error = %v", tt.filename, err)

		}

	}

}

func testValidateFileNameCustom(t *testing.T) {

	opts := security.Options{

		BlacklistFiles: []string{"temp*"},

		WhitelistFiles: []string{"important.txt"},
	}

	if err := security.ValidateFileName("temp_file.txt", opts); err == nil {

		t.Error("expected error for temp*")

	}

	if err := security.ValidateFileName("important.txt", opts); err != nil {

		t.Errorf("important.txt should be whitelisted: %v", err)

	}

}

func testValidateDirPath(t *testing.T) {

	opts := security.Options{Verbose: true}

	tests := []struct {
		path string

		wantErr bool
	}{

		{"safe/file.txt", false},

		{".git/file.txt", true},

		{".ssh/file.txt", true},
	}

	for _, tt := range tests {

		err := security.ValidateDirPath(tt.path, opts)

		if (err != nil) != tt.wantErr {

			t.Errorf("ValidateDirPath(%s) error = %v", tt.path, err)

		}

	}

}
