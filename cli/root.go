// Package cli assembles the papl command tree.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tamnd/papl-cli/papl"
)

// Build metadata, set via -ldflags at release time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// exit codes.
const (
	exitError  = 1
	exitUsage  = 2
	exitNoData = 3
	exitNet    = 5
)

// globalFlags holds parsed global flags shared by subcommands.
type globalFlags struct {
	DBPath    string
	StatePath string
	ExportDir string
	BaseURL   string
	DelayMs   int
	TimeoutS  int
	Workers   int
}

// Root builds and returns the root cobra command for the papl binary.
func Root() *cobra.Command {
	gf := &globalFlags{}
	cfg := papl.DefaultConfig()

	root := &cobra.Command{
		Use:   "papl",
		Short: "Programming and Programming Languages (PAPL) textbook archiver",
		Long: `papl crawls the PAPL textbook at papl.cs.brown.edu and stores every
chapter locally as SQLite + Markdown. PAPL covers programming language theory
and implementation using the Pyret language.

Pipeline:
  1. seed          -- parse table of contents, enqueue all chapter URLs
  2. crawl         -- fetch each chapter, convert to Markdown, store in DB
  3. export        -- write all chapters to Markdown files
  4. info          -- show DB and queue stats`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&gf.DBPath, "db", cfg.DBPath, "Path to SQLite database")
	root.PersistentFlags().StringVar(&gf.StatePath, "state", cfg.StatePath, "Path to crawl-queue database")
	root.PersistentFlags().StringVar(&gf.ExportDir, "export-dir", cfg.ExportDir, "Markdown export directory")
	root.PersistentFlags().StringVar(&gf.BaseURL, "base-url", cfg.BaseURL, "PAPL edition base URL")
	root.PersistentFlags().IntVar(&gf.DelayMs, "delay", int(cfg.Delay.Milliseconds()), "Delay between requests (ms)")
	root.PersistentFlags().IntVar(&gf.TimeoutS, "timeout", int(cfg.Timeout.Seconds()), "HTTP timeout (seconds)")
	root.PersistentFlags().IntVar(&gf.Workers, "workers", cfg.Workers, "Parallel chapter fetch workers")

	root.AddCommand(newSeedCmd(gf))
	root.AddCommand(newCrawlCmd(gf))
	root.AddCommand(newExportCmd(gf))
	root.AddCommand(newInfoCmd(gf))
	root.AddCommand(newQueueCmd(gf))
	root.AddCommand(newResetFailedCmd(gf))
	root.AddCommand(newVersionCmd())

	return root
}

// run is the entry point called from main.
func run(root *cobra.Command, args []string) int {
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	return 0
}
