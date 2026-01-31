package config

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	cfg := &Config{
		Algorithm:    "sha256",
		OutputFormat: "default",
		MinSize:      0,
		MaxSize:      -1, // No limit
		Jobs:         0,  // Auto-detection
	}

	// Initialize structured fields
	cfg.Input.MinSize = 0
	cfg.Input.MaxSize = -1
	cfg.Processing.Algorithm = "sha256"
	cfg.Processing.Jobs = 0
	cfg.Output.Format = "default"

	// Discovery defaults
	cfg.Discovery.InternalPath = "internal"
	cfg.Discovery.MainEntry = "cmd/chexum/main.go"
	cfg.Discovery.DocsPath = "docs"

	return cfg
}

var ValidOutputFormats = []string{
	"default",
	"verbose",
	"json",
	"jsonl",
	"plain",
	"csv",
}
var ValidAlgorithms = []string{"sha256", "md5", "sha1", "sha512", "blake2b"}
