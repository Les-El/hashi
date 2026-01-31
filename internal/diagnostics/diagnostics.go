package diagnostics

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Les-El/chexum/internal/color"
	"github.com/Les-El/chexum/internal/config"
	"github.com/Les-El/chexum/internal/console"
	"github.com/Les-El/chexum/internal/hash"
	"github.com/Les-El/chexum/internal/security"
)

// RunDiagnostics executes the diagnostics mode checks.
func RunDiagnostics(cfg *config.Config, streams *console.Streams) int {
	c := color.NewColorHandler()

	fmt.Fprintf(streams.Out, "%s\n", "Running Chexum System Diagnostics...")
	fmt.Fprintf(streams.Out, "--------------------------------------------------\n")

	// 1. Environment
	fmt.Fprintf(streams.Out, "%s System Information:\n", c.Blue("[INFO]"))
	fmt.Fprintf(streams.Out, "  OS/Arch:      %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(streams.Out, "  CPUs:         %d\n", runtime.NumCPU())
	fmt.Fprintf(streams.Out, "  Go Version:   %s\n", runtime.Version())

	u, err := user.Current()
	if err == nil {
		fmt.Fprintf(streams.Out, "  User:         %s (uid=%s)\n", u.Username, u.Uid)
	}

	wd, _ := os.Getwd()
	fmt.Fprintf(streams.Out, "  Working Dir:  %s\n", wd)
	fmt.Fprintf(streams.Out, "  Version:      %s\n", config.VersionText())
	fmt.Fprintf(streams.Out, "\n")

	// 2. Algorithm Check
	if checkAlgorithm(cfg.Algorithm, c, streams) {
		fmt.Fprintf(streams.Out, "%s Algorithm '%s' sanity check passed.\n", c.Green("[PASS]"), cfg.Algorithm)
	} else {
		fmt.Fprintf(streams.Out, "%s Algorithm '%s' check FAILED.\n", c.Red("[FAIL]"), cfg.Algorithm)
	}
	fmt.Fprintf(streams.Out, "\n")

	// 3. File Inspection
	if len(cfg.Files) > 0 {
		fmt.Fprintf(streams.Out, "%s Inspecting %d input arguments:\n", c.Blue("[DIAG]"), len(cfg.Files))
		for _, f := range cfg.Files {
			inspectFile(f, c, streams)
		}
	} else if len(cfg.Hashes) > 0 {
		fmt.Fprintf(streams.Out, "%s Inspecting %d hash arguments:\n", c.Blue("[DIAG]"), len(cfg.Hashes))
		for _, h := range cfg.Hashes {
			inspectHash(h, c, streams)
		}
	} else {
		fmt.Fprintf(streams.Out, "%s No file or hash arguments to inspect.\n", c.Yellow("[INFO]"))
	}

	return config.ExitSuccess
}

func checkAlgorithm(algo string, c *color.Handler, streams *console.Streams) bool {
	computer, err := hash.NewComputer(algo)
	if err != nil {
		fmt.Fprintf(streams.Out, "  Error initializing %s: %v\n", algo, err)
		return false
	}
	// Basic computation check
	hashStr, err := computer.ComputeReader(strings.NewReader("chexum"))
	if err != nil {
		fmt.Fprintf(streams.Out, "  Computation failed: %v\n", err)
		return false
	}
	if len(hashStr) == 0 {
		return false
	}
	return true
}

func inspectFile(path string, c *color.Handler, streams *console.Streams) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = "unknown"
	}

	sanitizedPath := security.SanitizeOutput(path)
	sanitizedAbs := security.SanitizeOutput(abs)

	fmt.Fprintf(streams.Out, "  Checking '%s':\n", sanitizedPath)
	fmt.Fprintf(streams.Out, "    - Absolute path: %s\n", sanitizedAbs)

	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(streams.Out, "    - Exists: %s\n", c.Red("NO"))
		// Check parent
		parent := filepath.Dir(abs)
		pInfo, pErr := os.Stat(parent)
		if pErr == nil && pInfo.IsDir() {
			fmt.Fprintf(streams.Out, "    - Parent directory exists: %s (%s)\n",
				c.Green("YES"), security.SanitizeOutput(parent))
			// Check write perms on parent?
		} else {
			fmt.Fprintf(streams.Out, "    - Parent directory exists: %s (%s)\n",
				c.Red("NO"), security.SanitizeOutput(parent))
		}
	} else {
		fmt.Fprintf(streams.Out, "    - Exists: %s\n", c.Green("YES"))
		fmt.Fprintf(streams.Out, "    - Size: %d bytes\n", info.Size())
		fmt.Fprintf(streams.Out, "    - Mode: %s\n", info.Mode())
		fmt.Fprintf(streams.Out, "    - IsDir: %v\n", info.IsDir())

		// Try to read first byte
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				fmt.Fprintf(streams.Out, "    - Readable: %s (%v)\n", c.Red("NO"), err)
			} else {
				fmt.Fprintf(streams.Out, "    - Readable: %s\n", c.Green("YES"))
				f.Close()
			}
		}
	}
}

func inspectHash(h string, c *color.Handler, streams *console.Streams) {
	fmt.Fprintf(streams.Out, "  Checking hash string '%s':\n", h)
	detects := hash.DetectHashAlgorithm(h)
	if len(detects) > 0 {
		fmt.Fprintf(streams.Out, "    - Valid format: %s\n", c.Green("YES"))
		fmt.Fprintf(streams.Out, "    - Possible algorithms: %s\n", strings.Join(detects, ", "))
	} else {
		fmt.Fprintf(streams.Out, "    - Valid format: %s\n", c.Red("NO"))
		fmt.Fprintf(streams.Out, "    - Length: %d\n", len(h))
	}
}
