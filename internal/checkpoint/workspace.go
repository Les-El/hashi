package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// Workspace manages a dedicated temporary area for application operations.
// It uses Afero for filesystem abstraction, allowing for in-memory or disk-based storage.
type Workspace struct {
	Fs    afero.Fs
	Root  string
	isMem bool
}

// NewWorkspace creates a new workspace. If useMem is true, it uses an in-memory filesystem.
// Otherwise, it creates a uniquely named directory in the system's temporary storage.
var NewWorkspace = func(useMem bool) (*Workspace, error) {
	if useMem {
		return &Workspace{
			Fs:    afero.NewMemMapFs(),
			isMem: true,
		}, nil
	}

	root, err := os.MkdirTemp("", "chexum-workspace-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace root: %w", err)
	}

	return &Workspace{
		Fs:   afero.NewOsFs(),
		Root: root,
	}, nil
}

// Path returns a path within the workspace.
func (w *Workspace) Path(elem ...string) string {
	if w.isMem {
		return filepath.Join(append([]string{"/"}, elem...)...)
	}
	return filepath.Join(append([]string{w.Root}, elem...)...)
}

// Cleanup removes all resources associated with the workspace.
func (w *Workspace) Cleanup() error {
	if w.isMem {
		// For memory FS, we just let the object be garbage collected or clear it
		w.Fs = nil
		return nil
	}

	if w.Root != "" {
		return os.RemoveAll(w.Root)
	}
	return nil
}

// WriteFile is a helper to write data to the workspace.
func (w *Workspace) WriteFile(filename string, data []byte) error {
	if strings.Contains(filename, "..") {
		return fmt.Errorf("path traversal not allowed: %s", filename)
	}
	path := w.Path(filename)
	dir := filepath.Dir(path)

	if err := w.Fs.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return afero.WriteFile(w.Fs, path, data, 0644)
}

// ReadFile is a helper to read data from the workspace.
func (w *Workspace) ReadFile(filename string) ([]byte, error) {
	if strings.Contains(filename, "..") {
		return nil, fmt.Errorf("path traversal not allowed: %s", filename)
	}
	return afero.ReadFile(w.Fs, w.Path(filename))
}
