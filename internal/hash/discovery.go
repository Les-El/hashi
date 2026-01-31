// Package hash provides hash computation and file discovery.
//
// DESIGN PRINCIPLE: Smart Traversal
// ---------------------------------
// Scanning large directory trees can be slow and resource-intensive.
// This package uses a depth-first traversal (filepath.Walk) combined
// with early-exit pruning for hidden directories and non-recursive
// requests. This ensures we only "stat" the files that actually
// matter to the user.
package hash

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DiscoveryOptions defines criteria for file discovery.
// It combines behavior flags (Recursive, Hidden) with filtering
// rules (Include, Exclude, Size, Date).
type DiscoveryOptions struct {
	Recursive      bool
	Hidden         bool
	Include        []string
	Exclude        []string
	MinSize        int64
	MaxSize        int64
	ModifiedAfter  time.Time
	ModifiedBefore time.Time
}

// DiscoverFiles finds all files in the given paths based on options.
//
// PROCESS:
// 1. If no paths provided, default to current directory (".").
// 2. Iterate through each root path.
// 3. Handle the special "-" stdin marker by passing it through.
// 4. Perform a recursive walk starting at each root.
// 5. Apply early-pruning and filters via handlePath.
func DiscoverFiles(paths []string, opts DiscoveryOptions) ([]string, error) {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var discovered []string
	for _, root := range paths {
		if root == "-" {
			// Special case: stdin marker is not a file path but a source signal.
			discovered = append(discovered, root)
			continue
		}

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				// We return errors immediately to fulfill the "Robustness" mandate.
				// If we can't access part of the tree, the user should know why.
				return err
			}
			return handlePath(path, root, info, opts, &discovered)
		})
		if err != nil {
			return nil, err
		}
	}
	return discovered, nil
}

// handlePath decides whether to include, skip, or descend into a path.
func handlePath(path, root string, info os.FileInfo, opts DiscoveryOptions, discovered *[]string) error {
	// Skip the root directory itself if it's not the current directory.
	if path == root && info.IsDir() && path != "." {
		return nil
	}

	// 1. Handle hidden files and directories.
	// We prune hidden directories early (filepath.SkipDir) to avoid
	// unnecessary traversal of large hidden trees like .git or .node_modules.
	if !opts.Hidden && isHidden(path, root) {
		if info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}

	// 2. Handle recursion and file types.
	// If recursion is disabled, we skip any directory that isn't the root.
	if info.IsDir() {
		if path != root && !opts.Recursive {
			return filepath.SkipDir
		}
		return nil
	}

	// 2b. Resource Hardening: Only process regular files.
	// This prevents chexum from hanging on /dev/zero, pipes, or other
	// "infinite" or blocking sources which could be used for DoS.
	if !info.Mode().IsRegular() {
		return nil
	}

	// 3. Apply user-defined filters (Name, Size, Date).
	if !passesFilters(info, path, opts) {
		return nil
	}

	*discovered = append(*discovered, path)
	return nil
}

// passesFilters applies the "AND" logic for all active filters.
// A file must pass EVERY active filter to be returned.
func passesFilters(info os.FileInfo, path string, opts DiscoveryOptions) bool {
	// Size filters: Checked first as they are extremely fast (metadata only).
	if opts.MinSize > 0 && info.Size() < opts.MinSize {
		return false
	}
	if opts.MaxSize != -1 && info.Size() > opts.MaxSize {
		return false
	}

	// Date filters: Uses modification time from filesystem metadata.
	if !opts.ModifiedAfter.IsZero() && info.ModTime().Before(opts.ModifiedAfter) {
		return false
	}
	if !opts.ModifiedBefore.IsZero() && info.ModTime().After(opts.ModifiedBefore) {
		return false
	}

	// Name filters: Pattern matching via globbing.
	name := filepath.Base(path)

	// Exclude patterns take absolute precedence.
	for _, pattern := range opts.Exclude {
		if matched, _ := filepath.Match(pattern, name); matched {
			return false
		}
	}

	// If include patterns exist, the file MUST match at least one.
	if len(opts.Include) > 0 {
		for _, pattern := range opts.Include {
			if matched, _ := filepath.Match(pattern, name); matched {
				return true
			}
		}
		return false
	}

	return true
}

// isHidden checks if any part of the path is hidden (starts with a dot).
func isHidden(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		// Fallback to simple base name check if relationship is ambiguous.
		return strings.HasPrefix(filepath.Base(path), ".")
	}

	// We split the path and check each component. A file is considered
	// hidden if it OR any of its parent directories are hidden.
	parts := strings.Split(rel, string(filepath.Separator))
	for _, part := range parts {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return true
		}
	}
	return false
}
