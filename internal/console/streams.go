// Package console handles the "Global Split Streams" architecture for chexum.
//
// It enforces the strict separation of Data (stdout) and Context (stderr),
// and implements the "Tee" model for simultaneous file output.
package console

import (
	"io"
	"os"

	"github.com/Les-El/chexum/internal/config"
)

// Streams holds the configured output streams.
type Streams struct {
	Out io.Writer // DATA: Result output (stdout + output file)
	Err io.Writer // CONTEXT: Logs, progress, errors (stderr + log file)
}

// InitStreams initializes the application streams based on the configuration.
// It sets up the "Tee" writers for file output and logging.
//
// Returns:
//   - *Streams: The initialized streams
//   - func(): A cleanup function to close opened files (MUST be called)
//   - error: If initialization fails
func InitStreams(cfg *config.Config) (*Streams, func(), error) {
	var outWriters []io.Writer
	var errWriters []io.Writer
	var filesToClose []io.Closer

	// 1. Standard Streams (Always active as the base)
	outWriters = append(outWriters, os.Stdout)
	errWriters = append(errWriters, os.Stderr)

	manager := NewOutputManager(cfg, os.Stdin)

	// 2. Output File (Tee stdout)
	if cfg.OutputFile != "" {
		f, err := manager.OpenOutputFile(cfg.OutputFile, cfg.Append, cfg.Force)
		if err != nil {
			return nil, nil, err
		}
		if f != nil {
			outWriters = append(outWriters, f)
			filesToClose = append(filesToClose, f)
		}
	}

	// 3. Log File (Tee stderr)
	if cfg.LogFile != "" {
		// Log files always append by convention
		f, err := manager.OpenOutputFile(cfg.LogFile, true, true)
		if err != nil {
			return nil, nil, err
		}
		if f != nil {
			errWriters = append(errWriters, f)
			filesToClose = append(filesToClose, f)
		}
	}

	streams := &Streams{
		Out: io.MultiWriter(outWriters...),
		Err: io.MultiWriter(errWriters...),
	}

	cleanup := func() {
		for _, f := range filesToClose {
			f.Close()
		}
	}

	return streams, cleanup, nil
}
