package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Les-El/chexum/internal/security"
)

// CleanupPattern defines a file pattern for cleanup.
// It uses shell-style globbing (filepath.Match) to identify targets.
type CleanupPattern struct {
	// Pattern is the glob pattern to match against filenames in the base directory.
	Pattern string `json:"pattern" toml:"pattern"`
	// Description provides a human-readable explanation of what this pattern targets.
	Description string `json:"description" toml:"description"`
	// Enabled allows toggling individual patterns without removing them from configuration.
	Enabled bool `json:"enabled" toml:"enabled"`
}

// CleanupConfig defines configuration for the cleanup manager.
// It allows users to override defaults and provide custom behaviors.
type CleanupConfig struct {
	// StorageThreshold is the percentage of storage usage that triggers a cleanup suggestion.
	StorageThreshold float64 `json:"storage_threshold" toml:"storage_threshold"`
	// MaxRetentionDays is the maximum age for temporary files before they are considered stale.
	MaxRetentionDays int `json:"max_retention_days" toml:"max_retention_days"`
	// CustomPatterns allows adding project-specific cleanup rules.
	CustomPatterns []CleanupPattern `json:"custom_patterns" toml:"custom_patterns"`
	// ExcludePatterns allows preventing specific files from being cleaned up even if they match.
	ExcludePatterns []string `json:"exclude_patterns" toml:"exclude_patterns"`
}

// CleanupManager handles temporary file cleanup operations.
// It is designed to be reusable across different components of the chexum tool.
type CleanupManager struct {
	verbose    bool
	dryRun     bool
	patterns   []CleanupPattern
	config     CleanupConfig
	baseDir    string
	workspaces []*Workspace
}

// NewCleanupManager creates a new cleanup manager with default patterns.
// By default, it targets common Go build artifacts and chexum-specific temporary files.
func NewCleanupManager(verbose bool) *CleanupManager {
	cm := &CleanupManager{
		verbose: verbose,
		baseDir: os.TempDir(),
		patterns: []CleanupPattern{
			{Pattern: "chexum-*", Description: "Chexum temporary files", Enabled: true},
			{Pattern: "checkpoint-*", Description: "Checkpoint temporary files", Enabled: true},
			{Pattern: "test-*", Description: "Test temporary files", Enabled: true},
			{Pattern: "*.tmp", Description: "Generic temporary files", Enabled: true},
		},
		workspaces: make([]*Workspace, 0),
	}
	return cm
}

// RegisterWorkspace adds a workspace to the manager's tracking list.
func (c *CleanupManager) RegisterWorkspace(ws *Workspace) {
	c.workspaces = append(c.workspaces, ws)
}

// SetDryRun enables or disables dry-run mode.
// In dry-run mode, no files are actually removed from the file system.
func (c *CleanupManager) SetDryRun(enabled bool) {
	c.dryRun = enabled
}

// SetBaseDir sets the directory where cleanup operations are performed.
// This is primarily used for testing purposes.
func (c *CleanupManager) SetBaseDir(dir string) {
	c.baseDir = dir
}

// AddCustomPattern adds a new cleanup pattern to the manager.
// Patterns added here will be processed alongside default patterns.
func (c *CleanupManager) AddCustomPattern(pattern, description string) {
	c.patterns = append(c.patterns, CleanupPattern{
		Pattern:     pattern,
		Description: description,
		Enabled:     true,
	})
}

// LoadConfig loads cleanup configuration from a TOML file.
// It merges custom patterns from the file into the manager's active patterns.
func (c *CleanupManager) LoadConfig(path string) error {
	var config struct {
		Cleanup CleanupConfig `toml:"cleanup"`
	}
	if _, err := toml.DecodeFile(path, &config); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to decode config file: %w", err)
	}
	c.config = config.Cleanup
	for _, p := range c.config.CustomPatterns {
		if p.Enabled {
			c.AddCustomPattern(p.Pattern, p.Description)
		}
	}
	return nil
}

// ValidatePatterns checks if any patterns (including custom and exclude) are invalid.
// It uses filepath.Match to verify the syntax of each pattern.
func (c *CleanupManager) ValidatePatterns() error {
	for _, p := range c.patterns {
		if _, err := filepath.Match(p.Pattern, "test"); err != nil {
			return fmt.Errorf("invalid pattern %q: %w", p.Pattern, err)
		}
	}
	for _, p := range c.config.ExcludePatterns {
		if _, err := filepath.Match(p, "test"); err != nil {
			return fmt.Errorf("invalid exclude pattern %q: %w", p, err)
		}
	}
	return nil
}

// CleanupResult contains detailed information about a cleanup operation's results.
type CleanupResult struct {
	// FilesRemoved is the number of individual files deleted.
	FilesRemoved int
	// DirsRemoved is the number of directories deleted (including all their contents).
	DirsRemoved int
	// SpaceFreed is the total size in bytes of all removed items.
	SpaceFreed int64
	// Errors contains a list of any error messages encountered during the operation.
	Errors []string
	// Duration is the total time taken to perform the cleanup.
	Duration time.Duration
	// StorageUsageBefore is the storage usage percentage before cleanup started.
	StorageUsageBefore float64
	// StorageUsageAfter is the storage usage percentage after cleanup finished.
	StorageUsageAfter float64
	// DryRun indicates whether this was a simulated operation.
	DryRun bool
}

// CleanupTemporaryFiles removes temporary files matching the manager's patterns and disposes of tracked workspaces.
// If dry-run mode is enabled, it identifies targets and calculates potential savings without deleting anything.
func (c *CleanupManager) CleanupTemporaryFiles() (*CleanupResult, error) {
	start := time.Now()
	result := &CleanupResult{
		DryRun: c.dryRun,
	}

	// Get storage usage before cleanup
	result.StorageUsageBefore = c.getStorageUsage()

	if c.verbose {
		c.logStart(result.StorageUsageBefore)
	}

	// 1. Process tracked workspaces
	for _, ws := range c.workspaces {
		c.processWorkspace(ws, result)
	}
	// Clear workspaces list after processing
	if !c.dryRun {
		c.workspaces = make([]*Workspace, 0)
	}

	// 2. Process legacy patterns in base directory
	tmpDir := c.baseDir
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s directory: %w", tmpDir, err)
	}

	for _, entry := range entries {
		c.processEntry(entry, result)
	}

	// Get storage usage after cleanup
	if c.dryRun {
		result.StorageUsageAfter = result.StorageUsageBefore
	} else {
		result.StorageUsageAfter = c.getStorageUsage()
	}
	result.Duration = time.Since(start)

	if c.verbose {
		c.logResult(result)
	}

	return result, nil
}

func (c *CleanupManager) processWorkspace(ws *Workspace, result *CleanupResult) {
	if ws.isMem || ws.Root == "" {
		return
	}

	if size, err := c.getDirSize(ws.Root); err == nil {
		result.SpaceFreed += size
	}

	if c.verbose {
		sanitizedRoot := security.SanitizeOutput(ws.Root)
		if c.dryRun {
			fmt.Printf("Would remove workspace: %s\n", sanitizedRoot)
		} else {
			fmt.Printf("Removing workspace: %s\n", sanitizedRoot)
		}
	}

	if !c.dryRun {
		if err := ws.Cleanup(); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to cleanup workspace %s: %v", ws.Root, err))
		} else {
			result.DirsRemoved++
		}
	} else {
		result.DirsRemoved++
	}
}

func (c *CleanupManager) logStart(usage float64) {
	if c.dryRun {
		fmt.Printf("Starting temporary file cleanup (DRY RUN)...\n")
	} else {
		fmt.Printf("Starting temporary file cleanup...\n")
	}
	fmt.Printf("Storage usage before cleanup: %.1f%%\n", usage)
}

func (c *CleanupManager) processEntry(entry os.DirEntry, result *CleanupResult) {
	name := entry.Name()
	filePath := filepath.Join(c.baseDir, name)

	if !c.shouldClean(name) {
		return
	}

	if info, err := entry.Info(); err == nil {
		if entry.IsDir() {
			if size, err := c.getDirSize(filePath); err == nil {
				result.SpaceFreed += size
			}
		} else {
			result.SpaceFreed += info.Size()
		}
	}

	if c.verbose {
		sanitizedPath := security.SanitizeOutput(filePath)
		if c.dryRun {
			fmt.Printf("Would remove: %s\n", sanitizedPath)
		} else {
			fmt.Printf("Removing: %s\n", sanitizedPath)
		}
	}

	if !c.dryRun {
		if err := os.RemoveAll(filePath); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to remove %s: %v", filePath, err))
		} else {
			if entry.IsDir() {
				result.DirsRemoved++
			} else {
				result.FilesRemoved++
			}
		}
	} else {
		if entry.IsDir() {
			result.DirsRemoved++
		} else {
			result.FilesRemoved++
		}
	}
}

func (c *CleanupManager) shouldClean(name string) bool {
	matched := false
	for _, p := range c.patterns {
		if !p.Enabled {
			continue
		}
		if m, _ := filepath.Match(p.Pattern, name); m {
			matched = true
			break
		}
	}

	if !matched {
		return false
	}

	for _, exclude := range c.config.ExcludePatterns {
		if m, _ := filepath.Match(exclude, name); m {
			return false
		}
	}

	return true
}

func (c *CleanupManager) logResult(result *CleanupResult) {
	fmt.Printf("Cleanup completed in %v\n", result.Duration)
	if c.dryRun {
		fmt.Printf("Files that would be removed: %d, Directories that would be removed: %d\n", result.FilesRemoved, result.DirsRemoved)
		fmt.Printf("Estimated space freed: %s\n", c.formatBytes(result.SpaceFreed))
	} else {
		fmt.Printf("Files removed: %d, Directories removed: %d\n", result.FilesRemoved, result.DirsRemoved)
		fmt.Printf("Space freed: %s\n", c.formatBytes(result.SpaceFreed))
		fmt.Printf("Storage usage after cleanup: %.1f%%\n", result.StorageUsageAfter)
	}
	if len(result.Errors) > 0 {
		fmt.Printf("Errors encountered: %d\n", len(result.Errors))
	}
}

// getDirSize calculates the total size of a directory recursively.
func (c *CleanupManager) getDirSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// formatBytes formats a byte count into a human-readable string (e.g., "1.5 MB").
func (c *CleanupManager) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// PreviewCleanup returns what would be removed without actually removing anything.
// It is a shorthand for enabling dry-run mode, running cleanup, and then restoring the original mode.
func (c *CleanupManager) PreviewCleanup() (*CleanupResult, error) {
	oldDryRun := c.dryRun
	c.dryRun = true
	defer func() { c.dryRun = oldDryRun }()
	return c.CleanupTemporaryFiles()
}

// CleanupOnExit performs cleanup and reports results to stdout.
// This is typically called at the end of a command or when an error occurs.
func (c *CleanupManager) CleanupOnExit() error {
	result, err := c.CleanupTemporaryFiles()
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	// Always report summary, even in non-verbose mode
	if result.DryRun {
		fmt.Printf("\n=== Cleanup Preview (DRY RUN) ===\n")
		fmt.Printf("Files that would be removed: %d\n", result.FilesRemoved)
		fmt.Printf("Directories that would be removed: %d\n", result.DirsRemoved)
		fmt.Printf("Estimated space freed: %s\n", c.formatBytes(result.SpaceFreed))
	} else {
		fmt.Printf("\n=== Cleanup Summary ===\n")
		fmt.Printf("Files removed: %d\n", result.FilesRemoved)
		fmt.Printf("Directories removed: %d\n", result.DirsRemoved)
		fmt.Printf("Space freed: %s\n", c.formatBytes(result.SpaceFreed))
		fmt.Printf("Storage usage: %.1f%% → %.1f%%\n", result.StorageUsageBefore, result.StorageUsageAfter)
	}
	fmt.Printf("Duration: %v\n", result.Duration)

	if len(result.Errors) > 0 {
		fmt.Printf("Errors: %d\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	return nil
}

// CheckStorageUsage checks if current storage usage exceeds the specified threshold.
// Returns true if usage is above threshold, along with the current usage percentage.
func (c *CleanupManager) CheckStorageUsage(threshold float64) (bool, float64) {
	usage := c.getStorageUsage()
	return usage > threshold, usage
}
