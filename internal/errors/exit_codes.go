// Package errors provides error handling and exit code logic.
package errors

import (
	"os"

	"github.com/Les-El/chexum/internal/config"
	"github.com/Les-El/chexum/internal/hash"
)

// ExitCode constants are defined in internal/config/config.go to avoid circular dependencies
// and because they are part of the application's configuration contract.

// DetermineDiscoveryExitCode determines the exit code for errors during file discovery.
func DetermineDiscoveryExitCode(err error) int {
	if os.IsNotExist(err) {
		return config.ExitFileNotFound
	}
	if os.IsPermission(err) {
		return config.ExitPermissionErr
	}
	return config.ExitPartialFailure
}

// DetermineExitCode determines the final exit code based on the operation results and configuration.
func DetermineExitCode(cfg *config.Config, result *hash.Result) int {
	if code := handleFailureExitCode(result); code != config.ExitSuccess {
		return code
	}
	return handleMatchExitCode(cfg, result)
}

func handleFailureExitCode(result *hash.Result) int {
	if len(result.Errors) == 0 {
		return config.ExitSuccess
	}

	// If all files failed, return the most specific error if possible
	if len(result.Errors) == len(result.Entries) && len(result.Entries) > 0 {
		groups := GroupErrors(result.Errors)
		if _, ok := groups[ErrorTypePermission]; ok {
			return config.ExitPermissionErr
		}
		if _, ok := groups[ErrorTypeFileNotFound]; ok {
			return config.ExitFileNotFound
		}
	}
	return config.ExitPartialFailure
}

func handleMatchExitCode(cfg *config.Config, result *hash.Result) int {
	if cfg.AnyMatch || cfg.MatchRequired {
		if len(result.Matches) > 0 || len(result.PoolMatches) > 0 {
			return config.ExitSuccess
		}
		return config.ExitNoMatches
	}

	if cfg.AllMatch {
		if len(cfg.Files) == 0 {
			return config.ExitSuccess
		}
		matchedFiles := make(map[string]bool)
		for _, m := range result.PoolMatches {
			matchedFiles[m.FilePath] = true
		}
		if len(result.Matches) == 1 && len(result.Unmatched) == 0 {
			return config.ExitSuccess
		}
		if len(matchedFiles) == len(cfg.Files) {
			return config.ExitSuccess
		}
		return config.ExitNoMatches
	}

	return config.ExitSuccess
}

