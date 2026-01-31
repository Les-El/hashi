package config

import (
	"fmt"
	"strings"

	"github.com/Les-El/chexum/internal/conflict"
	"github.com/Les-El/chexum/internal/security"
)

// ValidateOutputFormat checks if the provided format string is supported.
func ValidateOutputFormat(format string) error {
	for _, valid := range ValidOutputFormats {
		if format == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid output format %q: must be one of %s", format, strings.Join(ValidOutputFormats, ", "))
}

// ValidateAlgorithm checks if the provided algorithm string is supported.
func ValidateAlgorithm(algorithm string) error {
	for _, valid := range ValidAlgorithms {
		if algorithm == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid algorithm %q: must be one of %s", algorithm, strings.Join(ValidAlgorithms, ", "))
}

// ValidateConfig validates the configuration and returns an error if invalid.
func ValidateConfig(cfg *Config) ([]conflict.Warning, error) {
	warnings := make([]conflict.Warning, 0)

	opts := security.Options{
		Verbose:        cfg.Verbose,
		BlacklistFiles: cfg.BlacklistFiles,
		BlacklistDirs:  cfg.BlacklistDirs,
		WhitelistFiles: cfg.WhitelistFiles,
		WhitelistDirs:  cfg.WhitelistDirs,
	}

	// 1. Security validation of inputs
	if err := security.ValidateInputs(cfg.Files, cfg.Hashes, opts); err != nil {
		return warnings, err
	}

	// 2. Format and algorithm validation
	if err := falgValidate(cfg); err != nil {
		return warnings, err
	}

	// 3. Size and date validation
	if err := validateConstraints(cfg); err != nil {
		return warnings, err
	}

	// 4. Output path validation
	if err := validateAllOutputPaths(cfg); err != nil {
		return warnings, err
	}

	return warnings, nil
}

func falgValidate(cfg *Config) error {
	if err := ValidateOutputFormat(cfg.OutputFormat); err != nil {
		return err
	}
	return ValidateAlgorithm(cfg.Algorithm)
}

func validateConstraints(cfg *Config) error {
	if cfg.MinSize < 0 {
		return fmt.Errorf("min-size must be non-negative, got %d", cfg.MinSize)
	}
	if cfg.MaxSize != -1 && cfg.MaxSize < 0 {
		return fmt.Errorf("max-size must be non-negative or -1 (no limit), got %d", cfg.MaxSize)
	}
	if cfg.MaxSize != -1 && cfg.MinSize > cfg.MaxSize {
		return fmt.Errorf("min-size (%d) cannot be greater than max-size (%d)", cfg.MinSize, cfg.MaxSize)
	}

	if !cfg.ModifiedAfter.IsZero() && !cfg.ModifiedBefore.IsZero() {
		if cfg.ModifiedAfter.After(cfg.ModifiedBefore) {
			return fmt.Errorf("modified-after (%s) cannot be later than modified-before (%s)",
				cfg.ModifiedAfter.Format("2006-01-02"), cfg.ModifiedBefore.Format("2006-01-02"))
		}
	}
	return nil
}

func validateAllOutputPaths(cfg *Config) error {
	if err := validateOutputPath(cfg.OutputFile, cfg); err != nil {
		return fmt.Errorf("output file: %w", err)
	}
	if err := validateOutputPath(cfg.LogFile, cfg); err != nil {
		return fmt.Errorf("log file: %w", err)
	}
	if err := validateOutputPath(cfg.LogJSON, cfg); err != nil {
		return fmt.Errorf("JSON log file: %w", err)
	}
	return nil
}

// validateOutputPath validates that an output path is safe.
func validateOutputPath(path string, cfg *Config) error {
	opts := security.Options{
		Verbose:        cfg.Verbose,
		BlacklistFiles: cfg.BlacklistFiles,
		BlacklistDirs:  cfg.BlacklistDirs,
		WhitelistFiles: cfg.WhitelistFiles,
		WhitelistDirs:  cfg.WhitelistDirs,
	}
	return security.ValidateOutputPath(path, opts)
}
