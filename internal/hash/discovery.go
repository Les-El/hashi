// Package hash provides hash computation and file discovery.
package hash

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DiscoveryOptions defines criteria for file discovery.
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
// If paths is empty, it uses the current directory.
func DiscoverFiles(paths []string, opts DiscoveryOptions) ([]string, error) {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var discovered []string

	for _, root := range paths {
		if root == "-" {
			discovered = append(discovered, root)
			continue
		}

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip the root if it's a directory we're walking
			if path == root && info.IsDir() && path != "." {
				// But we still want to enter it
				return nil
			}

			// 1. Handle hidden files
			if !opts.Hidden && isHidden(path, root) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// 2. Handle directories
			if info.IsDir() {
				if path != root && !opts.Recursive {
					return filepath.SkipDir
				}
				return nil
			}

			// 3. Apply filters
			
			// Size filters
			if opts.MinSize > 0 && info.Size() < opts.MinSize {
				return nil
			}
			if opts.MaxSize != -1 && info.Size() > opts.MaxSize {
				return nil
			}

			// Date filters
			if !opts.ModifiedAfter.IsZero() && info.ModTime().Before(opts.ModifiedAfter) {
				return nil
			}
			if !opts.ModifiedBefore.IsZero() && info.ModTime().After(opts.ModifiedBefore) {
				return nil
			}

			// Name/Glob filters
			name := filepath.Base(path)
			
			// Exclude patterns (checked first)
			for _, pattern := range opts.Exclude {
				if matched, _ := filepath.Match(pattern, name); matched {
					return nil
				}
			}

			// Include patterns
			if len(opts.Include) > 0 {
				anyMatch := false
				for _, pattern := range opts.Include {
					if matched, _ := filepath.Match(pattern, name); matched {
						anyMatch = true
						break
					}
				}
				if !anyMatch {
					return nil
				}
			}
			
			discovered = append(discovered, path)
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return discovered, nil
}

// isHidden checks if a file or directory is hidden.
func isHidden(path, root string) bool {
	// Simple check: starts with dot
	// We check the base name of the current path element
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return strings.HasPrefix(filepath.Base(path), ".")
	}
	
	parts := strings.Split(rel, string(filepath.Separator))
	for _, part := range parts {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return true
		}
	}
	return false
}
