package cli

import (
	"time"

	"github.com/tamnd/papl-cli/papl"
)

// buildConfig constructs a papl.Config from global flags.
func buildConfig(gf *globalFlags) papl.Config {
	cfg := papl.DefaultConfig()
	if gf.DBPath != "" {
		cfg.DBPath = gf.DBPath
	}
	if gf.StatePath != "" {
		cfg.StatePath = gf.StatePath
	}
	if gf.ExportDir != "" {
		cfg.ExportDir = gf.ExportDir
	}
	if gf.BaseURL != "" {
		cfg.BaseURL = gf.BaseURL
	}
	if gf.DelayMs > 0 {
		cfg.Delay = time.Duration(gf.DelayMs) * time.Millisecond
	}
	if gf.TimeoutS > 0 {
		cfg.Timeout = time.Duration(gf.TimeoutS) * time.Second
	}
	if gf.Workers > 0 {
		cfg.Workers = gf.Workers
	}
	return cfg
}
