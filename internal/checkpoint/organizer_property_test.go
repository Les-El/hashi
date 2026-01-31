package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Les-El/chexum/internal/testutil"
)

// Property 4: Directory Structure Consistency
// **Validates: Requirements 2.1, 2.3, 2.4**
func TestProperty_DirectoryStructureConsistency(t *testing.T) {
	f := func(snapshotName string) bool {
		if snapshotName == "" {
			return true
		}

		tmpDir, cleanup := testutil.TempDir(t)
		defer cleanup()

		organizer := NewOrganizer(tmpDir)

		// Setup latest
		latestDir := filepath.Join(tmpDir, "active", "latest")
		os.MkdirAll(latestDir, 0755)
		os.WriteFile(filepath.Join(latestDir, "test.md"), []byte("test"), 0644)

		if err := organizer.CreateSnapshot(snapshotName); err != nil {
			return false
		}

		// Check structure
		snapshotPath := filepath.Join(tmpDir, "active", "snapshots", snapshotName)
		if _, err := os.Stat(snapshotPath); err != nil {
			return false
		}
		if _, err := os.Stat(filepath.Join(snapshotPath, "test.md")); err != nil {
			return false
		}

		return true
	}

	testutil.CheckProperty(t, f)
}

// Property 6: Snapshot Count Invariant
// **Validates: Requirements 2.5**
func TestProperty_SnapshotCountInvariant(t *testing.T) {
	f := func(numSnapshots int, maxActive int) bool {
		if numSnapshots < 0 {
			numSnapshots = -numSnapshots
			if numSnapshots < 0 {
				numSnapshots = 0
			}
		}
		numSnapshots = numSnapshots % 30
		
		if maxActive < 0 {
			maxActive = -maxActive
			if maxActive < 0 {
				maxActive = 5
			}
		}
		maxActive = maxActive%10 + 1

		tmpDir, cleanup := testutil.TempDir(t)
		defer cleanup()

		organizer := NewOrganizer(tmpDir)
		snapshotsDir := filepath.Join(tmpDir, "active", "snapshots")
		os.MkdirAll(snapshotsDir, 0755)

		for i := 0; i < numSnapshots; i++ {
			name := fmt.Sprintf("snapshot_%03d", i)
			os.MkdirAll(filepath.Join(snapshotsDir, name), 0755)
		}

		if err := organizer.ArchiveOldSnapshots(maxActive); err != nil {
			return false
		}

		entries, _ := os.ReadDir(snapshotsDir)
		expected := numSnapshots
		if numSnapshots > maxActive {
			expected = maxActive
		}

		return len(entries) == expected
	}

	testutil.CheckProperty(t, f)
}

// Property 5: Checkpoint Archival Behavior
// **Validates: Requirements 2.2**
func TestProperty_CheckpointArchivalBehavior(t *testing.T) {
	f := func(numSnapshots int) bool {
		if numSnapshots < 0 {
			numSnapshots = -numSnapshots
			if numSnapshots < 0 {
				numSnapshots = 0
			}
		}
		numSnapshots = numSnapshots%20 + 5 // At least 5

		tmpDir, cleanup := testutil.TempDir(t)
		defer cleanup()

		organizer := NewOrganizer(tmpDir)
		snapshotsDir := filepath.Join(tmpDir, "active", "snapshots")
		os.MkdirAll(snapshotsDir, 0755)

		for i := 0; i < numSnapshots; i++ {
			name := fmt.Sprintf("snapshot_%03d", i)
			os.MkdirAll(filepath.Join(snapshotsDir, name), 0755)
		}

		maxActive := 3
		if err := organizer.ArchiveOldSnapshots(maxActive); err != nil {
			return false
		}

		// Verify archive exists and contains the difference
		archiveBase := filepath.Join(tmpDir, "archive")
		foundInArchive := 0
		err := filepath.Walk(archiveBase, func(path string, info os.FileInfo, err error) error {
			if err == nil && info.IsDir() && filepath.Base(filepath.Dir(path)) != "archive" && path != archiveBase {
				// This is a snapshot directory inside a monthly directory
				if filepath.Base(filepath.Dir(filepath.Dir(path))) == "archive" {
					foundInArchive++
				}
			}
			return nil
		})
		if err != nil {
			return false
		}

		expectedArchived := numSnapshots - maxActive
		return foundInArchive == expectedArchived
	}

	testutil.CheckProperty(t, f)
}
