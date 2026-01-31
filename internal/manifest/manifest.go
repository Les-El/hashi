// Package manifest handles loading, saving, and comparing file manifests.
package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Les-El/chexum/internal/hash"
)

// Manifest represents a snapshot of file hashes and metadata.
type Manifest struct {
	Version   int          `json:"version"`
	Algorithm string       `json:"algorithm"`
	Created   time.Time    `json:"created"`
	Files     []FileRecord `json:"files"`
}

// FileRecord represents metadata for a single file in the manifest.
type FileRecord struct {
	Path  string    `json:"path"`
	Size  int64     `json:"size"`
	Mtime time.Time `json:"mtime"`
	Hash  string    `json:"hash"`
}

// New creates a new Manifest.
func New(algorithm string, entries []hash.Entry) *Manifest {
	m := &Manifest{
		Version:   1,
		Algorithm: algorithm,
		Created:   time.Now(),
		Files:     make([]FileRecord, 0, len(entries)),
	}

	for _, e := range entries {
		if e.Error == nil {
			m.Files = append(m.Files, FileRecord{
				Path:  e.Original,
				Size:  e.Size,
				Mtime: e.ModTime,
				Hash:  e.Hash,
			})
		}
	}

	return m
}

// Load reads a manifest from a file.
func Load(path string) (*Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var m Manifest
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}

	return &m, nil
}

// Save writes a manifest to a file using atomic write.
func Save(m *Manifest, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tempPath := path + ".tmp"
	f, err := os.Create(tempPath)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(m); err != nil {
		f.Close()
		os.Remove(tempPath)
		return err
	}

	if err := f.Close(); err != nil {
		os.Remove(tempPath)
		return err
	}

	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return err
	}

	return nil
}

// ChangeStatus represents the status of a file compared to a manifest.
type ChangeStatus int

const (
	Unchanged ChangeStatus = iota
	Modified
	Added
	Missing
)

// GetChangedFiles compares current files against the manifest.
func (m *Manifest) GetChangedFiles(currentFiles []string) ([]string, error) {
	manifestMap := make(map[string]FileRecord)
	for _, r := range m.Files {
		manifestMap[r.Path] = r
	}

	var changed []string
	for _, path := range currentFiles {
		info, err := os.Stat(path)
		if err != nil {
			// If file is missing now, it's changed (gone)
			changed = append(changed, path)
			continue
		}

		if info.IsDir() {
			continue
		}

		record, exists := manifestMap[path]
		if !exists {
			changed = append(changed, path)
			continue
		}

		// Change detection based on size and mtime
		if info.Size() != record.Size || !info.ModTime().Equal(record.Mtime) {
			changed = append(changed, path)
		}
	}

	return changed, nil
}
