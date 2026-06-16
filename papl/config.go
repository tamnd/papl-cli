package papl

import (
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultBaseURL is the root of the PAPL 2020 edition.
	DefaultBaseURL = "https://papl.cs.brown.edu/2020/"

	// DefaultDelay is the default wait between requests.
	DefaultDelay = 500 * time.Millisecond

	// DefaultWorkers is the default number of parallel chapter fetchers.
	DefaultWorkers = 4

	// DefaultTimeout is the default HTTP timeout.
	DefaultTimeout = 30 * time.Second
)

// Config holds constructor parameters for the CLI.
type Config struct {
	DBPath    string
	StatePath string
	ExportDir string
	BaseURL   string
	Workers   int
	Delay     time.Duration
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults rooted in $HOME/data/papl.
func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, "data", "papl")
	return Config{
		DBPath:    filepath.Join(dataDir, "papl.db"),
		StatePath: filepath.Join(dataDir, "state.db"),
		ExportDir: filepath.Join(dataDir, "export"),
		BaseURL:   DefaultBaseURL,
		Workers:   DefaultWorkers,
		Delay:     DefaultDelay,
		Timeout:   DefaultTimeout,
	}
}
