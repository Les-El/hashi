package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	organizer := NewOrganizer(tmpDir)

	// Create "latest" directory with some files
	latestDir := filepath.Join(tmpDir, "active", "latest")
	if err := os.MkdirAll(latestDir, 0755); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(latestDir, "report.md")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create snapshot
	snapshotName := "test_snapshot"
	if err := organizer.CreateSnapshot(snapshotName); err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	// Verify file moved
	snapshotFile := filepath.Join(tmpDir, "active", "snapshots", snapshotName, "report.md")
	if _, err := os.Stat(snapshotFile); os.IsNotExist(err) {
		t.Errorf("Expected snapshot file %s to exist", snapshotFile)
	}

	// Verify latest is empty
	entries, _ := os.ReadDir(latestDir)
	if len(entries) != 0 {
		t.Errorf("Expected latest directory to be empty, got %d entries", len(entries))
	}
}

func TestArchiveOldSnapshots(t *testing.T) {
	tmpDir := t.TempDir()
	organizer := NewOrganizer(tmpDir)

	snapshotsDir := filepath.Join(tmpDir, "active", "snapshots")
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create 10 snapshots
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("snapshot_2026-01-25_10-00-%02d", i)
		dir := filepath.Join(snapshotsDir, name)
		os.MkdirAll(dir, 0755)
	}

	// Maintain 5 active snapshots
	if err := organizer.ArchiveOldSnapshots(5); err != nil {
		t.Fatalf("ArchiveOldSnapshots failed: %v", err)
	}

	// Verify 5 left in active
	entries, _ := os.ReadDir(snapshotsDir)
	if len(entries) != 5 {
		t.Errorf("Expected 5 active snapshots, got %d", len(entries))
	}

	// Verify some moved to archive
	// Archive dir depends on current year-month
	archiveBase := filepath.Join(tmpDir, "archive")
	archiveEntries, err := os.ReadDir(archiveBase)
	if err != nil {
		t.Fatalf("Failed to read archive base: %v", err)
	}
	if len(archiveEntries) == 0 {
		t.Error("Expected at least one monthly archive directory")
	}
}

func TestNewOrganizer(t *testing.T) {
	o := NewOrganizer(".")
	if o.rootDir != "." {
		t.Errorf("expected ., got %s", o.rootDir)
	}
}

func TestGetActiveSnapshots(t *testing.T) {
	tmpDir := t.TempDir()
	organizer := NewOrganizer(tmpDir)

	snapshotsDir := filepath.Join(tmpDir, "active", "snapshots")
	os.MkdirAll(filepath.Join(snapshotsDir, "snap1"), 0755)
	os.MkdirAll(filepath.Join(snapshotsDir, "snap2"), 0755)
	os.WriteFile(filepath.Join(snapshotsDir, "notadir.txt"), []byte("junk"), 0644)

	infos, err := organizer.GetActiveSnapshots()
	if err != nil {
		t.Fatalf("GetActiveSnapshots failed: %v", err)
	}

	if len(infos) != 2 {
		t.Errorf("Expected 2 snapshots, got %d", len(infos))
	}

	t.Run("NonExistent", func(t *testing.T) {
		o := NewOrganizer(filepath.Join(tmpDir, "missing"))
		infos, err := o.GetActiveSnapshots()
		if err != nil {
			t.Fatalf("GetActiveSnapshots failed on missing dir: %v", err)
		}
		if len(infos) != 0 {
			t.Errorf("Expected 0 snapshots, got %d", len(infos))
		}
	})
}

func TestCleanupArchives(t *testing.T) {
	tmpDir := t.TempDir()
	organizer := NewOrganizer(tmpDir)

	archiveDir := filepath.Join(tmpDir, "archive")
	os.MkdirAll(filepath.Join(archiveDir, "2020-01"), 0755)
	os.MkdirAll(filepath.Join(archiveDir, "2025-01"), 0755)
	os.MkdirAll(filepath.Join(archiveDir, "invalid-name"), 0755)

	if err := organizer.CleanupArchives(1); err != nil {
		t.Errorf("CleanupArchives failed: %v", err)
	}

	// 2020-01 should be gone, 2025-01 should remain (assuming test runs in 2026), invalid-name should remain
	if _, err := os.Stat(filepath.Join(archiveDir, "2020-01")); !os.IsNotExist(err) {
		t.Error("Expected 2020-01 archive to be removed")
	}

	if _, err := os.Stat(filepath.Join(archiveDir, "invalid-name")); os.IsNotExist(err) {
		t.Error("Expected invalid-name archive to remain")
	}
}

func TestOrganizer_SnapshotError(t *testing.T) {
	tmpDir := t.TempDir()
	organizer := NewOrganizer(tmpDir)

	// Latest dir doesn't exist
	if err := organizer.CreateSnapshot("fail"); err == nil {
		t.Errorf("Expected failure when latest dir missing, got nil")
	}
}
