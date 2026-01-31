package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// CheckpointOrganizer manages the organization and archival of checkpoint artifacts.
type CheckpointOrganizer interface {
	// CreateSnapshot creates a new snapshot of current checkpoint artifacts.
	CreateSnapshot(name string) error
	// ArchiveOldSnapshots moves older snapshots to the archive directory.
	ArchiveOldSnapshots(maxActive int) error
	// CleanupArchives removes archives older than the specified retention period.
	CleanupArchives(retentionMonths int) error
	// GetActiveSnapshots returns information about snapshots in the active directory.
	GetActiveSnapshots() ([]SnapshotInfo, error)
}

// SnapshotStatus represents the state of a snapshot.
type SnapshotStatus string

const (
	StatusActive   SnapshotStatus = "active"
	StatusArchived SnapshotStatus = "archived"
)

// SnapshotInfo provides details about a checkpoint snapshot.
type SnapshotInfo struct {
	Name      string         `json:"name"`
	Timestamp time.Time      `json:"timestamp"`
	Size      int64          `json:"size"`
	Status    SnapshotStatus `json:"status"`
	Path      string         `json:"path"`
}

// Organizer handles the physical organization of checkpoint files.
type Organizer struct {
	rootDir string
}

// NewOrganizer creates a new checkpoint organizer.
func NewOrganizer(rootDir string) *Organizer {
	return &Organizer{
		rootDir: rootDir,
	}
}

// CreateSnapshot organizes current findings into a timestamped directory.
func (o *Organizer) CreateSnapshot(name string) error {
	activeDir := filepath.Join(o.rootDir, "active")
	latestDir := filepath.Join(activeDir, "latest")
	snapshotsBaseDir := filepath.Join(activeDir, "snapshots")

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	if name == "" {
		name = "snapshot_" + timestamp
	}

	// Security: Sanitize name to prevent path injection
	cleanName := filepath.Base(name)
	snapshotDir := filepath.Join(snapshotsBaseDir, cleanName)

	// Security: Double check that we are still within snapshotsBaseDir
	if rel, err := filepath.Rel(snapshotsBaseDir, snapshotDir); err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("invalid snapshot name: %s", name)
	}

	// Verify latest directory exists and is not empty
	if _, err := os.Stat(latestDir); err != nil {
		return fmt.Errorf("cannot create snapshot: latest findings directory missing: %w", err)
	}

	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Move files from latest to the new snapshot
	entries, err := os.ReadDir(latestDir)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("cannot create snapshot: latest findings directory is empty")
	}

	for _, entry := range entries {
		oldPath := filepath.Join(latestDir, entry.Name())
		newPath := filepath.Join(snapshotDir, entry.Name())
		// Logic: Use copy instead of rename to avoid emptying latest/
		if err := o.copyFile(oldPath, newPath); err != nil {
			return err
		}
	}

	return nil
}

// ArchiveOldSnapshots moves snapshots beyond maxActive to the archive.
func (o *Organizer) ArchiveOldSnapshots(maxActive int) error {
	snapshotsDir := filepath.Join(o.rootDir, "active", "snapshots")
	archiveDir := filepath.Join(o.rootDir, "archive", time.Now().Format("2006-01"))

	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var snapshots []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() {
			snapshots = append(snapshots, entry)
		}
	}

	if len(snapshots) <= maxActive {
		return nil
	}

	// Sort by name (which includes timestamp)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Name() < snapshots[j].Name()
	})

	// Archive the oldest ones
	toArchive := snapshots[:len(snapshots)-maxActive]

	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return err
	}

	for _, entry := range toArchive {
		oldPath := filepath.Join(snapshotsDir, entry.Name())
		newPath := filepath.Join(archiveDir, entry.Name())
		if err := os.Rename(oldPath, newPath); err != nil {
			return err
		}
	}

	return nil
}

// GetActiveSnapshots returns a list of current snapshots.
func (o *Organizer) GetActiveSnapshots() ([]SnapshotInfo, error) {
	snapshotsDir := filepath.Join(o.rootDir, "active", "snapshots")
	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SnapshotInfo{}, nil
		}
		return nil, err
	}

	var infos []SnapshotInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		infos = append(infos, SnapshotInfo{
			Name:      entry.Name(),
			Timestamp: info.ModTime(),
			Status:    StatusActive,
			Path:      filepath.Join(snapshotsDir, entry.Name()),
		})
	}

	return infos, nil
}

// CleanupArchives removes archive directories older than the specified retention period.
func (o *Organizer) CleanupArchives(retentionMonths int) error {
	archiveDir := filepath.Join(o.rootDir, "archive")
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	cutoff := time.Now().AddDate(0, -retentionMonths, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Monthly directories are named YYYY-MM
		t, err := time.Parse("2006-01", entry.Name())
		if err != nil {
			continue
		}

		if t.Before(cutoff) {
			if err := os.RemoveAll(filepath.Join(archiveDir, entry.Name())); err != nil {
				continue
			}
		}
	}
	return nil
}

// copyFile performs a simple file copy.
func (o *Organizer) copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}
