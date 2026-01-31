package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Les-El/chexum/internal/checkpoint"
)

func main() {

	if err := run(os.Args[1:], nil); err != nil {

		fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		os.Exit(1)

	}

}

func run(args []string, cm *checkpoint.CleanupManager) error {
	fs := flag.NewFlagSet("cleanup", flag.ContinueOnError)
	var (
		verbose   = fs.Bool("verbose", false, "Enable verbose output")
		dryRun    = fs.Bool("dry-run", false, "Show what would be cleaned without actually removing files")
		threshold = fs.Float64("threshold", 80.0, "Storage usage threshold percentage to trigger cleanup warning")
		force     = fs.Bool("force", false, "Force cleanup even if storage usage is below threshold")
	)

	if err := fs.Parse(args); err != nil {
		return err
	}

	if cm == nil {
		cm = checkpoint.NewCleanupManager(*verbose)
	}
	cm.SetDryRun(*dryRun)

	// Check current storage usage
	needsCleanup, usage := cm.CheckStorageUsage(*threshold)
	fmt.Printf("Current storage usage: %.1f%%\n", usage)

	if !needsCleanup && !*force && !*dryRun {
		fmt.Printf("Storage usage (%.1f%%) is below threshold (%.1f%%). Use -force to cleanup anyway.\n", usage, *threshold)
		return nil
	}

	if *dryRun {
		showDryRunInfo()
		return nil
	}

	if needsCleanup {
		fmt.Printf("Storage usage (%.1f%%) exceeds threshold (%.1f%%). Starting cleanup...\n", usage, *threshold)
	} else {
		fmt.Println("Force cleanup requested...")
	}

	if err := cm.CleanupOnExit(); err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	fmt.Println("Cleanup completed successfully!")
	return nil
}

func showDryRunInfo() {
	tmpDir := os.TempDir()
	fmt.Println("DRY RUN MODE - No files will be actually removed")
	fmt.Printf("Base directory: %s\n", tmpDir)
	fmt.Println("Would clean:")
	fmt.Printf("  - %s/go-build* directories\n", tmpDir)
	fmt.Printf("  - %s/chexum-* files\n", tmpDir)
	fmt.Printf("  - %s/checkpoint-* files\n", tmpDir)
	fmt.Printf("  - %s/test-* files\n", tmpDir)
	fmt.Printf("  - %s/*.tmp files\n", tmpDir)
}
