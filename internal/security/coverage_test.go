package security

import (
	"testing"
)

func TestValidateOutputPath_Coverage(t *testing.T) {
	t.Run("EmptyPath", func(t *testing.T) {
		if err := ValidateOutputPath("", Options{}); err != nil {
			t.Error(err)
		}
	})

	t.Run("InvalidExtension", func(t *testing.T) {
		if err := ValidateOutputPath("file.exe", Options{}); err == nil {
			t.Error("Expected error for .exe")
		}
	})

	t.Run("Traversal", func(t *testing.T) {
		if err := ValidateOutputPath("dir/../../etc/passwd", Options{}); err == nil {
			t.Error("Expected error for traversal")
		}
	})
}

func TestValidateFileName_Whitelist_Coverage(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		if err := ValidateFileName("", Options{}); err != nil {
			t.Error(err)
		}
	})

	t.Run("BlacklistedNoWhitelist", func(t *testing.T) {
		if err := ValidateFileName(".env", Options{}); err == nil {
			t.Error("Expected error for .env")
		}
	})

	t.Run("WhitelistMatch", func(t *testing.T) {
		opts := Options{
			WhitelistFiles: []string{"password_public.txt"},
		}
		if err := ValidateFileName("password_public.txt", opts); err != nil {
			t.Errorf("Expected success for whitelisted file, got %v", err)
		}
	})

	t.Run("WhitelistGlobMatch", func(t *testing.T) {
		opts := Options{
			WhitelistFiles: []string{"public_*"},
		}
		if err := ValidateFileName("public_password.txt", opts); err != nil {
			t.Errorf("Expected success for whitelisted file, got %v", err)
		}
	})
}

func TestValidateDirPath_Coverage(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		if err := ValidateDirPath("", Options{}); err != nil {
			t.Error(err)
		}
	})

	t.Run("ChexumConfig", func(t *testing.T) {
		if err := ValidateDirPath(".chexum/config", Options{}); err == nil {
			t.Error("Expected error for .chexum directory")
		}
	})

	t.Run("WhitelistDir", func(t *testing.T) {
		opts := Options{
			BlacklistDirs: []string{"secret"},
			WhitelistDirs: []string{"secret"},
		}
		if err := ValidateDirPath("secret/file.txt", opts); err != nil {
			t.Errorf("Expected success for whitelisted dir, got %v", err)
		}
	})
}

func TestValidateInputs_Coverage(t *testing.T) {
	err := ValidateInputs([]string{"-", "safe.txt"}, []string{"abc123"}, Options{})
	if err != nil {
		t.Error(err)
	}

	err = ValidateInputs(nil, []string{"NOT_HEX"}, Options{})
	if err == nil {
		t.Error("Expected error for non-hex hash")
	}
}

func TestResolveSafePath_Coverage(t *testing.T) {
	_, err := ResolveSafePath("safe/path")
	if err != nil {
		t.Error(err)
	}

	_, err = ResolveSafePath("../unsafe")
	if err == nil {
		t.Error("Expected error for traversal")
	}
}

func TestIsValidHex_Coverage(t *testing.T) {
	if isValidHex("") {
		t.Error("Expected false for empty string")
	}
}
